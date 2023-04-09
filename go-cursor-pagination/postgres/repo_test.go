package postgres_test

import (
	"context"
	"testing"

	"code.com/postgres"
	"code.com/product"
	"code.com/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestGetProducts_FirstPage_Limit5(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTable(t, db)

	store := postgres.NewStore(db)

	_, err := db.Exec(`
		INSERT INTO products(created_at, name)
		VALUES
			('2022-05-23 13:29:16', 'Shirt'),
			('2022-05-24 13:29:16', 'Polo'),
			('2022-05-25 13:29:16', 'T-Shirt'),
			('2022-05-26 13:29:16', 'Pants'),
			('2022-05-27 13:29:16', 'Socks'),
			('2022-05-28 13:29:16', 'Shoes'),
			('2022-05-29 13:29:16', 'Hat'),
			('2022-05-30 13:29:16', 'Glasses')
	`)
	if err != nil {
		t.Fatal(err)
	}

	cursors := postgres.Cursors{} // passing empty cursors, meaning we want the first page
	limit := 5

	pp, nextCursors, err := store.GetProducts(context.TODO(), cursors, limit)
	if err != nil {
		t.Fatal(err)
	}

	// we shouldn't get any `Prev` cursor because there is no previous page, we are on the
	// first page. We should get a cursor for the next page, which should be the creation data
	// of the last row we receive
	expectedCursors := postgres.Cursors{Next: "2022-05-26T13:29:16Z"}

	if diff := cmp.Diff(expectedCursors, nextCursors); diff != "" {
		t.Errorf("cursors are different (-want +got):\n%s", diff)
	}

	// we expect to receive the first 5 products sorted by creation date in descending order
	expectedProducts := []product.Product{
		{Name: "Glasses"},
		{Name: "Hat"},
		{Name: "Shoes"},
		{Name: "Socks"},
		{Name: "Pants"},
	}

	ignoreOpts := cmpopts.IgnoreFields(product.Product{}, "ID", "CreatedAt")
	if diff := cmp.Diff(expectedProducts, pp, ignoreOpts); diff != "" {
		t.Errorf("products are different (-want +got):\n%s", diff)
	}
}

func TestGetProducts(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTable(t, db)

	store := postgres.NewStore(db)

	_, err := db.Exec(`
		INSERT INTO products(created_at, name)
		VALUES
			('2022-05-23 13:29:16', 'Shirt'),
			('2022-05-24 13:29:16', 'Polo'),
			('2022-05-25 13:29:16', 'T-Shirt'),
			('2022-05-26 13:29:16', 'Pants'),
			('2022-05-27 13:29:16', 'Socks'),
			('2022-05-28 13:29:16', 'Shoes'),
			('2022-05-29 13:29:16', 'Hat'),
			('2022-05-30 13:29:16', 'Glasses')
	`)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		desc             string
		cursors          postgres.Cursors
		limit            int
		expectedProducts []product.Product
		expectedCursors  postgres.Cursors
	}{
		{
			desc:  "first page limit 5",
			limit: 5,
			expectedProducts: []product.Product{
				{Name: "Glasses"},
				{Name: "Hat"},
				{Name: "Shoes"},
				{Name: "Socks"},
				{Name: "Pants"},
			},
			expectedCursors: postgres.Cursors{
				Next: "2022-05-26T13:29:16Z",
			},
		},
		{
			desc:  "Next page limit 3",
			limit: 3,
			cursors: postgres.Cursors{
				Next: "2022-05-28T13:29:16Z",
			},
			expectedProducts: []product.Product{
				{Name: "Socks"},
				{Name: "Pants"},
				{Name: "T-Shirt"},
			},
			expectedCursors: postgres.Cursors{
				Prev: "2022-05-27T13:29:16Z",
				Next: "2022-05-25T13:29:16Z",
			},
		},
		{
			desc:  "going forward last page limit 3",
			limit: 3,
			cursors: postgres.Cursors{
				Next: "2022-05-26T13:29:16Z",
			},
			expectedProducts: []product.Product{
				{Name: "T-Shirt"},
				{Name: "Polo"},
				{Name: "Shirt"},
			},
			expectedCursors: postgres.Cursors{
				Prev: "2022-05-25T13:29:16Z",
			},
		},
		{
			desc:  "Go back first page limit 3",
			limit: 3,
			cursors: postgres.Cursors{
				Prev: "2022-05-22T13:29:16Z",
			},
			expectedProducts: []product.Product{
				{Name: "T-Shirt"},
				{Name: "Polo"},
				{Name: "Shirt"},
			},
			expectedCursors: postgres.Cursors{
				Prev: "2022-05-25T13:29:16Z",
			},
		},
		{
			desc:  "Go back limit 3",
			limit: 3,
			cursors: postgres.Cursors{
				Prev: "2022-05-24T13:29:16Z",
			},
			expectedProducts: []product.Product{
				{Name: "Socks"},
				{Name: "Pants"},
				{Name: "T-Shirt"},
			},
			expectedCursors: postgres.Cursors{
				Next: "2022-05-25T13:29:16Z",
				Prev: "2022-05-27T13:29:16Z",
			},
		},
		{
			desc:  "Go back last page limit 3",
			limit: 3,
			cursors: postgres.Cursors{
				Prev: "2022-05-27T13:29:16Z",
			},
			expectedProducts: []product.Product{
				{Name: "Glasses"},
				{Name: "Hat"},
				{Name: "Shoes"},
			},
			expectedCursors: postgres.Cursors{
				Next: "2022-05-28T13:29:16Z",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			pp, nextCursors, err := store.GetProducts(context.TODO(), tC.cursors, tC.limit)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tC.expectedCursors, nextCursors); diff != "" {
				t.Errorf("cursors are different (-want +got):\n%s", diff)
			}

			ignoreOpts := cmpopts.IgnoreFields(product.Product{}, "ID", "CreatedAt")
			if diff := cmp.Diff(tC.expectedProducts, pp, ignoreOpts); diff != "" {
				t.Errorf("products are different (-want +got):\n%s", diff)
			}
		})
	}
}
