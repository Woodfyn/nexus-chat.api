package transport

import (
	"fmt"
	"net/http"

	"github.com/Woodfyn/chat-api-backend-go/internal/transport/rest"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/Woodfyn/chat-api-backend-go/docs"

	"github.com/gorilla/mux"
)

type Api struct {
	rest *rest.Handler

	swaggerAddr string
}

func NewApi(rest *rest.Handler, swaggerAddr string) *Api {
	return &Api{
		rest: rest,

		swaggerAddr: swaggerAddr,
	}
}

func (a *Api) InitApi() *mux.Router {
	r := mux.NewRouter()

	// api
	api := r.PathPrefix("/api").Subrouter()
	{
		api.Use(a.rest.LoggingMiddleware)

		a.rest.InitRouter(api)
	}

	// swagger
	r.PathPrefix("/swagger").Handler(httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", a.swaggerAddr)),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
	)).Methods(http.MethodGet)

	return r
}
