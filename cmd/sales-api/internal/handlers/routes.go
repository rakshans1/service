package handlers

import (
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rakshans1/service/internal/mid"
	"github.com/rakshans1/service/internal/platform/auth"
	"github.com/rakshans1/service/internal/platform/web"
)

// API constructs an http.Handler will all apllication routes definde.
func API(db *sqlx.DB, log *log.Logger, authenticator *auth.Authenticator) http.Handler {
	app := web.NewApp(log, mid.Logger(log), mid.Errors(log), mid.Metrics())

	{
		c := Check{db: db}
		app.Handle(http.MethodGet, "/v1/health", c.Health)
	}

	{
		// Register user handlers.
		u := Users{db: db, authenticator: authenticator}
		app.Handle(http.MethodGet, "/v1/users/token", u.Token)
	}

	{

		p := Products{db: db, log: log}
		app.Handle(http.MethodGet, "/v1/products", p.List, mid.Authenticate(authenticator))
		app.Handle(http.MethodGet, "/v1/products/{id}", p.Retrive, mid.Authenticate(authenticator))
		app.Handle(http.MethodPost, "/v1/products", p.Create, mid.Authenticate(authenticator))
		app.Handle(http.MethodPut, "/v1/products/{id}", p.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
		app.Handle(http.MethodDelete, "/v1/products/{id}", p.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

		app.Handle(http.MethodPost, "/v1/products/{id}/sales", p.AddSale, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
		app.Handle(http.MethodGet, "/v1/products/{id}/sales", p.ListSales, mid.Authenticate(authenticator))

	}

	return app
}
