package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/telemetry"
	"github.com/swaggest/swgui/v5cdn"
	"github.com/valdirmendesdev/go-api-template/infra/ports/rest"
)

func main() {
	//Creates a base context for the app based on a root level and a signal that will be
	//notified if an interruption occur
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	//Cancel all child level contexts at the end of the function
	defer cancel()

	//Create Server
	restSrv := rest.NewServer()
	r := NewRouter(restSrv)

	//creates http server to expose api
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	fmt.Printf("Server is starting at %s\n", "8080")

	ch := make(chan error, 1)

	//Execute the server on a separate thread for graceful shutdown
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch: //server error
		panic(err)
	case <-ctx.Done(): //os interrupt
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_ = srv.Shutdown(timeout)
	}
}

func NewRouter(restSrv rest.Server) chi.Router {
	swagger, err := rest.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	//swagger.Servers = nil

	r := chi.NewRouter()

	//Enable CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(telemetry.Collector(telemetry.Config{
		AllowAny: true,
	}, []string{})) // path prefix filters records generic http request metrics

	// OpenAPI file descriptor
	r.Get("/openapi/schema.json", func(w http.ResponseWriter, r *http.Request) {
		swaggerJson, err := swagger.MarshalJSON()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(swaggerJson)
	})

	// Swagger UI
	r.Handle("/docs", v5cdn.New(
		"Template project Backend API",
		"/openapi/schema.json",
		"/api/openapi/",
	))

	r.Group(func(r chi.Router) {
		//r.Use(middleware.OapiRequestValidator(swagger))
		opts := rest.ChiServerOptions{
			BaseRouter: r,
		}

		r.Mount("/", rest.HandlerWithOptions(restSrv, opts))
	})

	return r
}
