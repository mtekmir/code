package postgres_test

import (
	"context"
	"testing"

	"code.com/postgres"
	"code.com/product"
	"code.com/test"
	"github.com/google/go-cmp/cmp"
)

func TestGetProduct(t *testing.T) {
	db := test.SetupTX(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)

	_, err := db.Exec(`INSERT INTO variations(id, name) VALUES(1, 'L')`)
	if err != nil {
		t.Fatalf("failed to insert variation. %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO products(id, barcode, name, price, quantity, variation_id) VALUES
			(1, '123456789', 'White Shirt', "100", 10, 1),
			(2, '234567890', 'Blue Scarf', "80", 10, NULL)
	`)
	if err != nil {
		t.Fatalf("failed to insert products. %v", err)
	}

	expectedProduct1 := &product.Product{
		ID:       1,
		Barcode:  "123456789",
		Name:     "White Shirt",
		Price:    "100",
		Quantity: 10,
		Variation: &product.Variation{
			ID:   1,
			Name: "L",
		},
	}

	p1, err := store.Get(context.TODO(), 1)
	if err != nil {
		t.Fatalf("failed to get product. %v", err)
	}

	if diff := cmp.Diff(expectedProduct1, p1); diff != "" {
		t.Errorf("products are different (-want +got):\n%s", diff)
	}
}
