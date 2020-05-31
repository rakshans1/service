package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
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
func (p *Products) List(w http.ResponseWriter, r *http.Request) {
	list, err := product.List(p.db)
	if err != nil {
		p.log.Println("listing products", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := web.Respond(w, list, http.StatusOK); err != nil {
		p.log.Println("encoding response", "error", err)
		return
	}
}

// Create decode the body of a request to create a new product. The full
// product with generated fields is sent back in the response
func (p *Products) Create(w http.ResponseWriter, r *http.Request) {
	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		p.log.Println("decoding product", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	prod, err := product.Create(p.db, np, time.Now())
	if err != nil {
		p.log.Println("creating product", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := web.Respond(w, &prod, http.StatusCreated); err != nil {
		p.log.Println("encoding response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Retrive finds a single product identified by an ID in the request URL.
func (p *Products) Retrive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	prod, err := product.Retrieve(p.db, id)
	if err != nil {
		p.log.Println("getting product", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return

	}

	if err := web.Respond(w, prod, http.StatusOK); err != nil {
		p.log.Println("encoding response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
