package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"code.com/product"
)

type ProductStore struct {
	db *sql.DB
}

func NewProductStore(db *sql.DB) ProductStore {
	return ProductStore{db: db}
}

func (store ProductStore) Get(ctx context.Context, ID product.ID) (*product.Product, error) {
	var p product.Product
	var vID sql.NullInt64
	var vName sql.NullString
	err := store.db.QueryRowContext(ctx, `
		SELECT p.id, barcode, p.name, price, quantity, v.id, v.name
		FROM products p
		LEFT JOIN variations v ON p.variation_id = v.id
		WHERE p.id = $1
	`, ID).Scan(&p.ID, &p.Barcode, &p.Name, &p.Price, &p.Quantity, &vID, &vName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("product with id %d not found", ID)
		}
		return nil, err
	}

	if vID.Valid {
		p.Variation = &product.Variation{ID: product.VariationID(vID.Int64), Name: vName.String}
	}

	return &p, nil
}
