package web

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

// Handler is the signature used by all application handlers in this service.
type Handler func(http.ResponseWriter, *http.Request) error

// App is the entrypoint into our application and what controls the context of
// each request. Feel free to add any configuration data/logic on this type
type App struct {
	log *log.Logger
	mux *chi.Mux
}

func NewApp(log *log.Logger) *App {
	return &App{
		log: log,
		mux: chi.NewRouter(),
	}
}

// Handle associates a handler function with an HTTP Method  and URL pattern.
func (a *App) Handle(method, url string, h Handler) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Call the handler and catch any propogated error.
		err := h(w, r)

		if err != nil {

			// Tell the client about the error.
			res := ErrorResponse{
				Error: err.Error(),
			}
			Respond(w, res, http.StatusInternalServerError)
		}
	}
	a.mux.MethodFunc(method, url, fn)
}

// ServeHTTP implements the http.Handler interface.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}
