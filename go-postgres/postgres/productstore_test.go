package postgres_test

import (
	"context"
	"testing"
	"time"

	"code.com/postgres"
	"code.com/product"
	"code.com/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestGetBrand(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)

	_, err := db.Exec(`INSERT INTO brands(id, name) VALUES(1, 'Nike')`)
	if err != nil {
		t.Fatalf("failed to insert brand. %v", err)
	}

	found, err := store.GetBrand(context.TODO(), 1)
	if err != nil {
		t.Fatal(err)
	}

	expectedBrand := product.Brand{ID: 1, Name: "Nike"}

	if diff := cmp.Diff(expectedBrand, found); diff != "" {
		t.Errorf("brands are different (-want +got):\n%s", diff)
	}
}

func TestInsertBrand(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)

	inserted, err := store.InsertBrand(context.TODO(), product.Brand{Name: "Abercrombie"})
	if err != nil {
		t.Fatal(err)
	}

	if inserted.ID != 1 {
		t.Errorf("expected brand id to be 1. Got %d", inserted.ID)
	}

	var found product.Brand
	err = db.QueryRowContext(context.TODO(), `
		SELECT id, name FROM brands WHERE id = 1
	`).Scan(&found.ID, &found.Name)
	if err != nil {
		t.Fatal(err)
	}

	expectedBrand := product.Brand{ID: 1, Name: "Abercrombie"}

	if diff := cmp.Diff(expectedBrand, found); diff != "" {
		t.Errorf("brands are different (-want +got):\n%s", diff)
	}
}

func TestGetProduct(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)

	_, err := db.Exec(`INSERT INTO brands(name) VALUES ('Nike')`)
	if err != nil {
		t.Fatalf("failed to insert brand. %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO products(created_at, name, price, brand_id) 
		VALUES
			('2022-05-25 13:29:09', 'White Shirt', 1000, 1),
			('2022-05-25 13:29:09', 'Blue Scarf', 800, NULL)
	`)
	if err != nil {
		t.Fatalf("failed to insert products. %v", err)
	}

	_, err = db.Exec(`INSERT INTO variations(name) VALUES ('L'), ('M')`)
	if err != nil {
		t.Fatalf("failed to insert variations. %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO product_variations(quantity, product_id, variation_id)
		VALUES (10, 1, 1), (20, 1, 2)
	`)
	if err != nil {
		t.Fatalf("failed to insert to product variations table. %v", err)
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", "2022-05-25 13:29:09")
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		desc            string
		id              product.ID
		expectedProduct product.Product
		error           bool
	}{
		{
			desc: "with variation and brand",
			id:   1,
			expectedProduct: product.Product{
				ID:        1,
				CreatedAt: createdAt,
				Name:      "White Shirt",
				Price:     1000,
				Brand:     &product.Brand{ID: 1, Name: "Nike"},
				Variations: []product.Variation{
					{ID: 1, Name: "L", Quantity: 10},
					{ID: 2, Name: "M", Quantity: 20},
				},
			},
		},
		{
			desc: "without variation and brand",
			id:   2,
			expectedProduct: product.Product{
				ID:        2,
				CreatedAt: createdAt,
				Name:      "Blue Scarf",
				Price:     800,
			},
		},
		{
			desc:            "not found",
			id:              3,
			expectedProduct: product.Product{},
			error:           true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			p1, err := store.GetProduct(context.TODO(), tC.id)
			if (err != nil) != tC.error {
				t.Fatalf(`productStore.Get() error = "%v". expected error = %v`, err, tC.error)
			}

			if diff := cmp.Diff(tC.expectedProduct, p1); diff != "" {
				t.Errorf("products are different (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetProducts(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)

	_, err := db.Exec(`INSERT INTO brands(name) VALUES ('Brand')`)
	if err != nil {
		t.Fatalf("failed to insert brand. %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO products(created_at, name, price, brand_id) 
		VALUES
			('2022-05-25 13:29:09', 'White Shirt', 1000, NULL),
			('2022-05-25 13:29:09', 'Blue Scarf', 800, 1),
			('2022-05-25 13:29:09', 'Red Scarf', 800, NULL),
			('2022-05-25 13:29:09', 'Black Scarf', 800, 1)
	`)
	if err != nil {
		t.Fatalf("failed to insert products. %v", err)
	}

	_, err = db.Exec(`INSERT INTO variations(name) VALUES ('L'), ('M')`)
	if err != nil {
		t.Fatalf("failed to insert variations. %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO product_variations(quantity, product_id, variation_id)
		VALUES (10, 1, 1), (20, 2, 1), (20, 2, 2)
	`)
	if err != nil {
		t.Fatalf("failed to insert to product variations table. %v", err)
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", "2022-05-25 13:29:09")
	if err != nil {
		t.Fatal(err)
	}

	expectedProducts := []product.Product{
		{
			ID:        1,
			CreatedAt: createdAt,
			Name:      "White Shirt",
			Price:     1000,
			Variations: []product.Variation{
				{ID: 1, Name: "L", Quantity: 10},
			},
		},
		{
			ID:        2,
			CreatedAt: createdAt,
			Name:      "Blue Scarf",
			Price:     800,
			Brand:     &product.Brand{ID: 1, Name: "Brand"},
			Variations: []product.Variation{
				{ID: 1, Name: "L", Quantity: 20},
				{ID: 2, Name: "M", Quantity: 20},
			},
		},
		{
			ID:        3,
			CreatedAt: createdAt,
			Name:      "Red Scarf",
			Price:     800,
		},
		{
			ID:        4,
			CreatedAt: createdAt,
			Name:      "Black Scarf",
			Price:     800,
			Brand:     &product.Brand{ID: 1, Name: "Brand"},
		},
	}

	pp, err := store.GetProducts(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	sortOpt := cmpopts.SortSlices(func(i, j product.Product) bool {
		return i.ID < j.ID
	})
	if diff := cmp.Diff(expectedProducts, pp, sortOpt); diff != "" {
		t.Errorf("product slices are different (-want +got):\n%s", diff)
	}
}

func TestInsertProduct(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)

	_, err := store.InsertBrand(context.TODO(), product.Brand{Name: "Brand"})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().Round(time.Microsecond)
	testCases := []struct {
		desc     string
		product  product.Product
		expected product.Product
	}{
		{
			desc: "with variation",
			product: product.Product{
				CreatedAt: now,
				Name:      "Product1",
				Price:     1000,
				Brand:     &product.Brand{ID: 1, Name: "Brand"},
			},
			expected: product.Product{
				ID:        1,
				CreatedAt: now,
				Name:      "Product1",
				Price:     1000,
				Brand:     &product.Brand{ID: 1, Name: "Brand"},
			},
		},
		{
			desc: "without variation",
			product: product.Product{
				CreatedAt: now,
				Name:      "Product2",
				Price:     2000,
			},
			expected: product.Product{
				ID:        2,
				CreatedAt: now,
				Name:      "Product2",
				Price:     2000,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			inserted, err := store.InsertProduct(context.TODO(), tC.product)
			if err != nil {
				t.Fatalf(err.Error())
			}

			if diff := cmp.Diff(tC.expected, inserted); diff != "" {
				t.Errorf("products are different (-want +got):\n%s", diff)
			}

			found, err := store.GetProduct(context.TODO(), tC.expected.ID)
			if err != nil {
				t.Fatalf(err.Error())
			}

			if diff := cmp.Diff(tC.expected, found); diff != "" {
				t.Errorf("products are different (-want +got):\n%s", diff)
			}
		})
	}
}

func TestInsertProductVariation(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)
	_, err := db.Exec(`
		INSERT INTO products(name, price, brand_id) 
		VALUES ('White Shirt', 1000, NULL)
	`)
	if err != nil {
		t.Fatalf("failed to insert product. %v", err)
	}

	v := product.Variation{Name: "XS", Quantity: 2}
	inserted, err := store.InsertProductVariation(context.TODO(), 1, v)
	if err != nil {
		t.Fatalf("failed insert product variation. %v", err)
	}

	if inserted.ID != 1 {
		t.Errorf("expected id to be 1. Got %d", inserted.ID)
	}

	var found product.Variation
	err = db.QueryRow(`
		SELECT v.id, name, pv.quantity FROM variations v
		JOIN product_variations pv ON pv.variation_id = v.id
		WHERE v.id = 1
	`).Scan(&found.ID, &found.Name, &found.Quantity)
	if err != nil {
		t.Fatalf("failed to get variation. %v", err)
	}

	expectedVariation := product.Variation{ID: 1, Name: "XS", Quantity: 2}

	if diff := cmp.Diff(expectedVariation, found); diff != "" {
		t.Errorf("products are different (-want +got):\n%s", diff)
	}
}

func TestDeleteProduct(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)

	_, err := db.Exec(`
		INSERT INTO products(name, price, brand_id) 
		VALUES
			('White Shirt', 1000, NULL),
			('Blue Scarf', 800, NULL),
			('Red Scarf', 800, NULL)
	`)
	if err != nil {
		t.Fatalf("failed to insert products. %v", err)
	}

	_, err = db.Exec(`INSERT INTO variations(name) VALUES ('L'), ('M')`)
	if err != nil {
		t.Fatalf("failed to insert variations. %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO product_variations(quantity, product_id, variation_id)
		VALUES (10, 1, 1), (20, 2, 1), (20, 2, 2)
	`)
	if err != nil {
		t.Fatalf("failed to insert to product variations table. %v", err)
	}

	testCases := []struct {
		desc             string
		idToDelete       product.ID
		expectedProducts []product.Product
	}{
		{
			desc:       "with variation",
			idToDelete: 1,
			expectedProducts: []product.Product{
				{
					ID:    2,
					Name:  "Blue Scarf",
					Price: 800,
					Variations: []product.Variation{
						{ID: 1, Name: "L", Quantity: 20},
						{ID: 2, Name: "M", Quantity: 20},
					},
				},
				{
					ID:    3,
					Name:  "Red Scarf",
					Price: 800,
				},
			},
		},
		{
			desc:       "with 2 variaions",
			idToDelete: 2,
			expectedProducts: []product.Product{
				{
					ID:    3,
					Name:  "Red Scarf",
					Price: 800,
				},
			},
		},
		{
			desc:             "without variation",
			idToDelete:       3,
			expectedProducts: []product.Product{},
		},
		{
			desc:             "not found",
			idToDelete:       4,
			expectedProducts: []product.Product{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			if err := store.DeleteProduct(context.TODO(), tC.idToDelete); err != nil {
				t.Fatal(err)
			}

			found, err := store.GetProducts(context.TODO())
			if err != nil {
				t.Fatal(err)
			}

			ignored := cmpopts.IgnoreFields(product.Product{}, "CreatedAt")
			if diff := cmp.Diff(tC.expectedProducts, found, ignored); diff != "" {
				t.Errorf("products are different (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBulkInsertProducts(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	_, err := db.Exec(`INSERT INTO brands(name) VALUES ('Nike')`)
	if err != nil {
		t.Fatalf("failed to insert brand. %v", err)
	}

	store := postgres.NewProductStore(db)

	createdAt := time.Now().Round(time.Microsecond)
	productsToInsert := []product.Product{
		{
			CreatedAt: createdAt,
			Name:      "Product1",
			Price:     1111,
		},
		{
			CreatedAt: createdAt,
			Name:      "Product2",
			Price:     2222,
			Brand:     &product.Brand{ID: 1, Name: "Nike"},
		},
		{
			Name:  "Product3",
			Price: 3333,
		},
		{
			Name:  "Product4",
			Price: 4444,
		},
	}

	expectedProducts := []product.Product{
		{
			ID:        1,
			CreatedAt: createdAt,
			Name:      "Product1",
			Price:     1111,
		},
		{
			ID:        2,
			CreatedAt: createdAt,
			Name:      "Product2",
			Price:     2222,
			Brand:     &product.Brand{ID: 1, Name: "Nike"},
		},
		{
			ID:    3,
			Name:  "Product3",
			Price: 3333,
		},
		{
			ID:    4,
			Name:  "Product4",
			Price: 4444,
		},
	}

	inserted, err := store.BulkInsertProducts(context.TODO(), productsToInsert)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expectedProducts, inserted); diff != "" {
		t.Errorf("products are different (-want +got):\n%s", diff)
	}

	found, err := store.GetProducts(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	sortOpt := cmpopts.SortSlices(func(i, j product.Product) bool {
		return i.ID < j.ID
	})
	if diff := cmp.Diff(expectedProducts, found, sortOpt); diff != "" {
		t.Errorf("products are different (-want +got):\n%s", diff)
	}
}

func TestSearchProducts(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)
	_, err := db.Exec(`
		INSERT INTO products(name, price) 
		VALUES
			('Shirt', 1000),
			('Scarf', 800),
			('T-Shirt', 1000),
			('Pants', 1000),
			('Socks', 1000),
			('Shoes', 1000),
			('Hat', 1000),
			('Glasses', 1000)
	`)
	if err != nil {
		t.Fatalf("failed to insert products. %v", err)
	}

	testCases := []struct {
		desc             string
		query            string
		expectedProducts []product.Product
	}{
		{
			desc:  "query:'s'",
			query: "s",
			expectedProducts: []product.Product{
				{Name: "Shirt", Price: 1000},
				{Name: "Scarf", Price: 800},
				{Name: "Socks", Price: 1000},
				{Name: "Shoes", Price: 1000},
			},
		},
		{
			desc:             "query:'asd'",
			query:            "asd",
			expectedProducts: []product.Product{},
		},
		{
			desc:  "query:'Ha'",
			query: "Ha",
			expectedProducts: []product.Product{
				{Name: "Hat", Price: 1000},
			},
		},
		{
			desc:  "query:'SH'",
			query: "SH",
			expectedProducts: []product.Product{
				{Name: "Shirt", Price: 1000},
				{Name: "Shoes", Price: 1000},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			pp, err := store.SearchProducts(context.TODO(), tC.query)
			if err != nil {
				t.Fatal(err)
			}

			ignoreOpts := cmpopts.IgnoreFields(product.Product{}, "ID", "CreatedAt")
			if diff := cmp.Diff(tC.expectedProducts, pp, ignoreOpts); diff != "" {
				t.Errorf("products are different (-want +got):\n%s", diff)
			}
		})
	}
}

func intToPtr(n int) *int { return &n }

func TestGetProductOffsetPagination(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)

	_, err := db.Exec(`
		INSERT INTO products(created_at, name, price)
		VALUES
			('2022-05-25 13:29:16', 'Shirt', 1000),
			('2022-05-25 13:29:15', 'Scarf', 800),
			('2022-05-25 13:29:14', 'T-Shirt', 1000),
			('2022-05-25 13:29:13', 'Pants', 1000),
			('2022-05-25 13:29:12', 'Socks', 1000),
			('2022-05-25 13:29:11', 'Shoes', 1000),
			('2022-05-25 13:29:10', 'Hat', 1000),
			('2022-05-25 13:29:09', 'Glasses', 1000)
	`)
	if err != nil {
		t.Fatalf("failed to insert products. %v", err)
	}

	expectedProducts := []product.Product{
		{Name: "Shirt", Price: 1000},
		{Name: "Scarf", Price: 800},
		{Name: "T-Shirt", Price: 1000},
		{Name: "Pants", Price: 1000},
		{Name: "Socks", Price: 1000},
		{Name: "Shoes", Price: 1000},
		{Name: "Hat", Price: 1000},
		{Name: "Glasses", Price: 1000},
	}

	testCases := []struct {
		desc             string
		page             *int
		rowsPerPage      *int
		expectedProducts []product.Product
		expectedTotal    int
	}{
		{
			desc:             "no pagination",
			expectedProducts: expectedProducts,
			expectedTotal:    8,
		},
		{
			desc:             "page 1 rowsPerPage 3",
			expectedProducts: expectedProducts[:3],
			page:             intToPtr(1),
			rowsPerPage:      intToPtr(3),
			expectedTotal:    8,
		},
		{
			desc:             "page 2 rowsPerPage 3",
			expectedProducts: expectedProducts[3:6],
			page:             intToPtr(2),
			rowsPerPage:      intToPtr(3),
			expectedTotal:    8,
		},
		{
			desc:             "page 2 rowsPerPage 3",
			expectedProducts: expectedProducts[6:8],
			page:             intToPtr(3),
			rowsPerPage:      intToPtr(3),
			expectedTotal:    8,
		},
		{
			desc:             "page 2 rowsPerPage 6",
			expectedProducts: expectedProducts[6:8],
			page:             intToPtr(2),
			rowsPerPage:      intToPtr(6),
			expectedTotal:    8,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			pp, c, err := store.GetProductsOffsetPagination(context.TODO(), tC.page, tC.rowsPerPage)
			if err != nil {
				t.Fatal(err)
			}

			if c != tC.expectedTotal {
				t.Errorf("Expected count to be %d. Got %d", tC.expectedTotal, c)
			}

			ignoreOpts := cmpopts.IgnoreFields(product.Product{}, "ID", "CreatedAt")
			if diff := cmp.Diff(tC.expectedProducts, pp, ignoreOpts); diff != "" {
				t.Errorf("products are different (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetProductsFilteredSorted(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)

	_, err := db.Exec(`
		INSERT INTO brands(id, name) 
		VALUES (1, 'Brand1'), (2, 'Brand2')
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO products(created_at, name, price, brand_id)
		VALUES
			('2022-05-25 13:29:16', 'Shirt', 5000, 1),
			('2022-05-25 13:29:15', 'Polo', 4000, 1),
			('2022-05-25 13:29:14', 'T-Shirt', 3000, 1),
			('2022-05-25 13:29:13', 'Pants', 8000, NULL),
			('2022-05-25 13:29:12', 'Socks', 1000, NULL),
			('2022-05-25 13:29:11', 'Shoes', 2000, 2),
			('2022-05-25 13:29:10', 'Hat', 500, 2),
			('2022-05-25 13:29:09', 'Glasses', 2500, 2)
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO variations(id, name)
		VALUES
			(1, 'L'), (2, 'M'), (3, 'S')
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO product_variations(product_id, variation_id, quantity)
		VALUES
			(1, 1, 1), (2, 2, 2), (2, 3, 3), (3, 1, 4), (4, 2, 5), (5, 3, 3)
	`)
	if err != nil {
		t.Fatal(err)
	}

	brand1 := &product.Brand{ID: 1, Name: "Brand1"}
	brand2 := &product.Brand{ID: 2, Name: "Brand2"}

	testCases := []struct {
		desc             string
		params           postgres.Params
		expectedProducts []product.Product
		expectedTotal    int
	}{
		{
			desc: "price between 5000 and 2000 order by price asc",
			params: postgres.Params{
				StartPrice:       intToPtr(2000),
				EndPrice:         intToPtr(5000),
				OrderBy:          postgres.OrderByPrice,
				OrderByDirection: postgres.OrderByDirectionASC,
			},
			expectedProducts: []product.Product{
				{Name: "Shoes", Price: 2000, Brand: brand2},
				{Name: "Glasses", Price: 2500, Brand: brand2},
				{
					Name: "T-Shirt", Price: 3000, Brand: brand1,
					Variations: []product.Variation{
						{ID: 1, Name: "L", Quantity: 4},
					},
				},
				{
					Name: "Polo", Price: 4000, Brand: brand1,
					Variations: []product.Variation{
						{ID: 2, Name: "M", Quantity: 2},
						{ID: 3, Name: "S", Quantity: 3},
					},
				},
				{
					Name: "Shirt", Price: 5000, Brand: brand1,
					Variations: []product.Variation{
						{ID: 1, Name: "L", Quantity: 1},
					},
				},
			},
			expectedTotal: 5,
		},
		{
			desc: "brandIds 1 and 2 order by quantity desc",
			params: postgres.Params{
				BrandIDs:         []product.BrandID{1, 2},
				OrderBy:          postgres.OrderByPrice,
				OrderByDirection: postgres.OrderByDirectionDESC,
			},
			expectedProducts: []product.Product{
				{
					Name: "Shirt", Price: 5000, Brand: brand1,
					Variations: []product.Variation{
						{ID: 1, Name: "L", Quantity: 1},
					},
				},
				{
					Name: "Polo", Price: 4000, Brand: brand1,
					Variations: []product.Variation{
						{ID: 2, Name: "M", Quantity: 2},
						{ID: 3, Name: "S", Quantity: 3},
					},
				},
				{
					Name: "T-Shirt", Price: 3000, Brand: brand1,
					Variations: []product.Variation{
						{ID: 1, Name: "L", Quantity: 4},
					},
				},
				{Name: "Glasses", Price: 2500, Brand: brand2},
				{Name: "Shoes", Price: 2000, Brand: brand2},
				{Name: "Hat", Price: 500, Brand: brand2},
			},
			expectedTotal: 6,
		},
		{
			desc: "filter and pagination together",
			params: postgres.Params{
				BrandIDs:         []product.BrandID{1, 2},
				OrderBy:          postgres.OrderByPrice,
				OrderByDirection: postgres.OrderByDirectionASC,
				Page:             intToPtr(2),
				Limit:            intToPtr(2),
			},
			expectedProducts: []product.Product{
				{Name: "Glasses", Price: 2500, Brand: brand2},
				{
					Name: "T-Shirt", Price: 3000, Brand: brand1,
					Variations: []product.Variation{
						{ID: 1, Name: "L", Quantity: 4},
					},
				},
			},
			expectedTotal: 6,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			pp, c, err := store.GetProductsFilteredSorted(context.TODO(), tC.params)
			if err != nil {
				t.Fatal(err)
			}

			if c != tC.expectedTotal {
				t.Errorf("Expected count to be %d. Got %d", tC.expectedTotal, c)
			}

			ignoreOpts := cmpopts.IgnoreFields(product.Product{}, "ID", "CreatedAt")
			if diff := cmp.Diff(tC.expectedProducts, pp, ignoreOpts); diff != "" {
				t.Errorf("products are different (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetProductsCursorPagination(t *testing.T) {
	db := test.SetupDB(t)
	test.CreateProductTables(t, db)

	store := postgres.NewProductStore(db)

	_, err := db.Exec(`
		INSERT INTO brands(id, name) 
		VALUES (1, 'Brand1'), (2, 'Brand2')
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO products(created_at, name, price, brand_id)
		VALUES
			('2022-05-23 13:29:16', 'Shirt', 5000, 1),
			('2022-05-24 13:29:16', 'Polo', 4000, 1),
			('2022-05-25 13:29:16', 'T-Shirt', 3000, 1),
			('2022-05-26 13:29:16', 'Pants', 8000, NULL),
			('2022-05-27 13:29:16', 'Socks', 1000, NULL),
			('2022-05-28 13:29:16', 'Shoes', 2000, 2),
			('2022-05-29 13:29:16', 'Hat', 500, 2),
			('2022-05-30 13:29:16', 'Glasses', 2500, 2)
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO variations(id, name)
		VALUES (1, 'L'), (2, 'M'), (3, 'S')
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO product_variations(product_id, variation_id, quantity)
		VALUES (1, 1, 1), (2, 2, 2), (2, 3, 3)
	`)
	if err != nil {
		t.Fatal(err)
	}

	brand1 := &product.Brand{ID: 1, Name: "Brand1"}
	brand2 := &product.Brand{ID: 2, Name: "Brand2"}

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
				{Name: "Glasses", Price: 2500, Brand: brand2},
				{Name: "Hat", Price: 500, Brand: brand2},
				{Name: "Shoes", Price: 2000, Brand: brand2},
				{Name: "Socks", Price: 1000},
				{Name: "Pants", Price: 8000},
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
				{Name: "Socks", Price: 1000},
				{Name: "Pants", Price: 8000},
				{Name: "T-Shirt", Price: 3000, Brand: brand1},
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
				{Name: "T-Shirt", Price: 3000, Brand: brand1},
				{
					Name: "Polo", Price: 4000, Brand: brand1,
					Variations: []product.Variation{
						{ID: 2, Name: "M", Quantity: 2},
						{ID: 3, Name: "S", Quantity: 3},
					},
				},
				{
					Name: "Shirt", Price: 5000, Brand: brand1,
					Variations: []product.Variation{
						{ID: 1, Name: "L", Quantity: 1},
					},
				},
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
				{Name: "T-Shirt", Price: 3000, Brand: brand1},
				{
					Name: "Polo", Price: 4000, Brand: brand1,
					Variations: []product.Variation{
						{ID: 2, Name: "M", Quantity: 2},
						{ID: 3, Name: "S", Quantity: 3},
					},
				},
				{
					Name: "Shirt", Price: 5000, Brand: brand1,
					Variations: []product.Variation{
						{ID: 1, Name: "L", Quantity: 1},
					},
				},
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
				{Name: "Socks", Price: 1000},
				{Name: "Pants", Price: 8000},
				{Name: "T-Shirt", Price: 3000, Brand: brand1},
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
				{Name: "Glasses", Price: 2500, Brand: brand2},
				{Name: "Hat", Price: 500, Brand: brand2},
				{Name: "Shoes", Price: 2000, Brand: brand2},
			},
			expectedCursors: postgres.Cursors{
				Next: "2022-05-28T13:29:16Z",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			pp, nextCursors, err := store.GetProductsCursorPagination(context.TODO(), tC.cursors, tC.limit)
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
