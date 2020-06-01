package handlers

import (
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rakshans1/service/internal/mid"
	"github.com/rakshans1/service/internal/platform/web"
)

// API constructs an http.Handler will all apllication routes definde.
func API(db *sqlx.DB, log *log.Logger) http.Handler {
	app := web.NewApp(log, mid.Errors(log), mid.Metrics())

	{
		c := Check{db: db}
		app.Handle(http.MethodGet, "/v1/health", c.Health)
	}

	{

		p := Products{db: db, log: log}
		app.Handle(http.MethodGet, "/v1/products", p.List)
		app.Handle(http.MethodGet, "/v1/products/{id}", p.Retrive)
		app.Handle(http.MethodPost, "/v1/products", p.Create)
		app.Handle(http.MethodPut, "/v1/products/{id}", p.Update)
		app.Handle(http.MethodDelete, "/v1/products/{id}", p.Delete)

		app.Handle(http.MethodPost, "/v1/products/{id}/sales", p.AddSale)
		app.Handle(http.MethodGet, "/v1/products/{id}/sales", p.ListSales)

	}

	return app
}
