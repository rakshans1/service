package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rakshans1/service/internal/platform/web"
	"github.com/rakshans1/service/internal/product"
)

// Products defines all of the handlers related to products. It holds the
// application state needed by the handler methods.
type Products struct {
	db  *sqlx.DB
	log *log.Logger
}

// List gets all products from the service layer and encodes them for the
// client response.
func (p *Products) List(w http.ResponseWriter, r *http.Request) error {
	list, err := product.List(p.db)
	if err != nil {
		return errors.Wrap(err, "getting product list")
	}

	return web.Respond(w, list, http.StatusOK)
}

// Create decode the body of a request to create a new product. The full
// product with generated fields is sent back in the response
func (p *Products) Create(w http.ResponseWriter, r *http.Request) error {
	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return errors.Wrap(err, "decoding new product")
	}

	prod, err := product.Create(p.db, np, time.Now())
	if err != nil {
		return errors.Wrap(err, "creating new product")
	}

	return web.Respond(w, &prod, http.StatusCreated)
}

// Retrive finds a single product identified by an ID in the request URL.
func (p *Products) Retrive(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	prod, err := product.Retrieve(p.db, id)
	if err != nil {
		switch err {
		case product.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)

		case product.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)

		default:
			return errors.Wrapf(err, "getting product %q", id)
		}
	}

	return web.Respond(w, prod, http.StatusOK)
}
