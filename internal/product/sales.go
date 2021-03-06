package product

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
)

// AddSale records a sales transaction for a single Product.
func AddSale(ctx context.Context, db *sqlx.DB, ns NewSale, productID string, now time.Time) (*Sale, error) {
	ctx, span := global.Tracer("service").Start(ctx, "internal.product.addsale")
	defer span.End()

	s := Sale{
		ID:          uuid.New().String(),
		ProductID:   productID,
		Quantity:    ns.Quantity,
		Paid:        ns.Paid,
		DateCreated: now,
	}

	const q = `INSERT into sales
		(sale_id, product_id,  quantity, paid, date_created)
		VALUES ($1,$2,$3,$4, $5)`

	_, err := db.ExecContext(ctx, q, s.ID, s.ProductID, s.Quantity, s.Paid, s.DateCreated)
	if err != nil {
		return nil, errors.Wrap(err, "inserting sale")
	}

	return &s, nil
}

// ListSales gives all Sales for a Product.
func ListSales(ctx context.Context, db *sqlx.DB, productID string) ([]Sale, error) {
	ctx, span := global.Tracer("service").Start(ctx, "internal.product.listsales")
	defer span.End()

	sales := []Sale{}

	const q = `SELECT * FROM sales WHERE product_id = $1`
	if err := db.SelectContext(ctx, &sales, q, productID); err != nil {
		return nil, errors.Wrap(err, "selecting sales")
	}
	return sales, nil
}
