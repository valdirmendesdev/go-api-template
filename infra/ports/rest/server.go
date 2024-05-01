//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=../../../api/openapi/generation-config.yaml ../../../api/openapi/schema.yaml
package rest

type Server struct {
}

func NewServer() Server {
	s := Server{}
	return s
}
