package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rakshans1/service/internal/platform/auth"
	"github.com/rakshans1/service/internal/platform/web"
	"github.com/rakshans1/service/internal/product"
	"go.opentelemetry.io/otel/api/global"
)

// Products defines all of the handlers related to products. It holds the
// application state needed by the handler methods.
type Products struct {
	db  *sqlx.DB
	log *log.Logger
}

// List gets all products from the service layer and encodes them for the
// client response.
func (p *Products) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.list")
	defer span.End()

	list, err := product.List(ctx, p.db)
	if err != nil {
		return errors.Wrap(err, "getting product list")
	}

	return web.Respond(ctx, w, list, http.StatusOK)
}

// Create decode the body of a request to create a new product. The full
// product with generated fields is sent back in the response
func (p *Products) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.products.create")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)

	if !ok {
		return errors.New("claims missing from context")
	}

	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return errors.Wrap(err, "decoding new product")
	}

	prod, err := product.Create(ctx, p.db, claims, np, time.Now())
	if err != nil {
		return errors.Wrap(err, "creating new product")
	}

	return web.Respond(ctx, w, &prod, http.StatusCreated)
}

// Retrive finds a single product identified by an ID in the request URL.
func (p *Products) Retrive(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.products.get")
	defer span.End()

	id := chi.URLParam(r, "id")

	prod, err := product.Get(ctx, p.db, id)
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

	return web.Respond(ctx, w, prod, http.StatusOK)
}

// Update decodes the body of a request to update an existing product. The ID
// of the product is part of the request URL.
func (p *Products) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.products.update")
	defer span.End()

	id := chi.URLParam(r, "id")

	var update product.UpdateProduct
	if err := web.Decode(r, &update); err != nil {
		return errors.Wrap(err, "decoding product update")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	if err := product.Update(ctx, p.db, claims, id, update, time.Now()); err != nil {
		switch err {
		case product.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case product.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "updating product %q", id)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes a single product identified by an ID in the request URL.
func (p *Products) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.products.delete")
	defer span.End()

	id := chi.URLParam(r, "id")

	if err := product.Delete(r.Context(), p.db, id); err != nil {
		switch err {
		case product.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "deleting product %q", id)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// AddSale creates a new Sale for a particular product. It looks for a JSON
// object in the request body. The full model is returned to the caller.
func (p *Products) AddSale(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.products.addsale")
	defer span.End()

	var ns product.NewSale
	if err := web.Decode(r, &ns); err != nil {
		return errors.Wrap(err, "decoding new sale")
	}

	productID := chi.URLParam(r, "id")

	sale, err := product.AddSale(ctx, p.db, ns, productID, time.Now())
	if err != nil {
		return errors.Wrap(err, "adding new sale")
	}

	return web.Respond(ctx, w, sale, http.StatusCreated)
}

// ListSales gets all sales for a particular product.
func (p *Products) ListSales(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.products.listsales")
	defer span.End()

	id := chi.URLParam(r, "id")

	list, err := product.ListSales(ctx, p.db, id)
	if err != nil {
		return errors.Wrap(err, "getting sales list")
	}

	return web.Respond(ctx, w, list, http.StatusOK)
}
