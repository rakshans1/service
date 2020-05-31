package handlers

import (
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rakshans1/service/internal/platform/web"
)

// API constructs an http.Handler will all apllication routes definde.
func API(db *sqlx.DB, log *log.Logger) http.Handler {
	app := web.NewApp(log)

	p := Products{db: db, log: log}
	app.Handle(http.MethodGet, "/v1/products", p.List)
	app.Handle(http.MethodGet, "/v1/products/{id}", p.Retrive)
	app.Handle(http.MethodPost, "/v1/products", p.Create)

	return app
}
