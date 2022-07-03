package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"code.com/config"
	"code.com/postgres"
	"code.com/product"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	conf := config.Parse()

	db, dbClose, err := postgres.Setup(conf.DatabaseURL)
	if err != nil {
		return err
	}
	defer dbClose()

	store := postgres.NewProductStore(db)

	type b struct {
		Limit int
		Prev  string
		Next  string
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var body b
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		pp, cursors, err := store.GetProductsCursorPagination(r.Context(), postgres.Cursors{Prev: body.Prev, Next: body.Next}, body.Limit)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		products := []product.Product{}
		for _, p := range pp {
			pr := &p
			pr.Variations = []product.Variation{}
			pr.Brand = nil
			products = append(products, *pr)
		}

		type res struct {
			Products []product.Product
			Prev     string
			Next     string
		}
		w.Header().Add("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(res{Products: products, Next: cursors.Next, Prev: cursors.Prev}); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

	})

	http.ListenAndServe(":8080", nil)

	// generateProducts(db)

	return nil
}

func generateProducts(db *sql.DB) {
	store := postgres.NewProductStore(db)

	brands := []*product.Brand{
		{Name: "Brand1"},
		{Name: "Brand2"},
		{Name: "Brand3"},
	}

	for _, b := range brands {
		r, err := store.InsertBrand(context.Background(), *b)
		if err != nil {
			log.Fatal(err)
		}
		b.ID = r.ID
	}

	now := time.Now().Round(time.Hour)
	day := time.Hour * 24
	products := []product.Product{}

	for i := 0; i < 50; i++ {
		products = append(products, product.Product{
			Name:      fmt.Sprintf("Product%d", i+1),
			CreatedAt: now.Add(time.Duration(-1*i) * day),
			Price:     100 * (i + 1),
			Brand:     brands[i%3],
		})
	}

	pp, err := store.BulkInsertProducts(context.Background(), products)
	if err != nil {
		log.Fatal(err)
	}

	for i, p := range pp {
		for j := 0; j < rand.Intn(4); j++ {
			v := product.Variation{
				Name:     fmt.Sprintf("Variation%d%d", i+1, j+1),
				Quantity: j,
			}
			_, err := store.InsertProductVariation(context.Background(), p.ID, v)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
