package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"code.com/product"
)

type ProductStore struct {
	db *sql.DB
}

func NewProductStore(db *sql.DB) ProductStore {
	return ProductStore{db: db}
}

func (store ProductStore) InsertBrand(ctx context.Context, v product.Brand) (product.Brand, error) {
	err := store.db.QueryRowContext(ctx, `
		INSERT INTO brands(name) VALUES($1) RETURNING id
	`, v.Name).Scan(&v.ID)
	if err != nil {
		return product.Brand{}, fmt.Errorf("failed to insert brand. %v", err)
	}

	return v, nil
}

func (store ProductStore) GetBrand(ctx context.Context, ID product.BrandID) (product.Brand, error) {
	var b product.Brand
	err := store.db.QueryRowContext(ctx, `
		SELECT id, name FROM brands WHERE id = $1
	`, ID).Scan(&b.ID, &b.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return product.Brand{}, fmt.Errorf("brand with id %d not found", ID)
		}
		return product.Brand{}, fmt.Errorf("failed to get brand. error: %v", err)
	}

	return b, nil
}

func (store ProductStore) GetProduct(ctx context.Context, ID product.ID) (product.Product, error) {
	var p product.Product

	rows, err := store.db.QueryContext(ctx, `
		SELECT p.id, created_at, p.name, price, v.id, v.name, pv.quantity, b.id, b.name
		FROM products p
		LEFT JOIN product_variations pv ON p.id = pv.product_id
		LEFT JOIN variations v ON pv.variation_id = v.id
		LEFT JOIN brands b ON p.brand_id = b.id
		WHERE p.id = $1
	`, ID)
	if err != nil {
		return product.Product{}, err
	}

	found := false
	for rows.Next() {
		found = true

		var (
			brandID       sql.NullInt64
			brandName     sql.NullString
			variationID   sql.NullInt64
			variationName sql.NullString
			quantity      sql.NullInt64
		)

		err := rows.Scan(&p.ID, &p.CreatedAt, &p.Name, &p.Price, &variationID, &variationName, &quantity, &brandID, &brandName)
		if err != nil {
			return product.Product{}, err
		}

		// Variation and brand is optional, so only include them if they exist

		if variationID.Valid {
			v := product.Variation{
				ID:       product.VariationID(variationID.Int64),
				Name:     variationName.String,
				Quantity: int(quantity.Int64),
			}
			p.Variations = append(p.Variations, v)
		}

		if brandID.Valid {
			p.Brand = &product.Brand{ID: product.BrandID(brandID.Int64), Name: brandName.String}
		}
	}

	if !found {
		return product.Product{}, fmt.Errorf("product with id %d not found", ID)
	}

	return p, nil
}

func (store ProductStore) GetProducts(ctx context.Context) ([]product.Product, error) {
	stmt := `
		SELECT p.id, created_at, p.name, price, v.id, v.name, pv.quantity, b.id, b.name
		FROM products p
		LEFT JOIN product_variations pv ON p.id = pv.product_id
		LEFT JOIN variations v ON pv.variation_id = v.id
		LEFT JOIN brands b ON p.brand_id = b.id
		ORDER BY created_at DESC
	`

	rows, err := store.db.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to query products. %v", err)
	}
	defer rows.Close()

	// We need to keep the products in a map, because we'll get
	// the same product result multiple times if it has multiple
	// variations.
	pp := map[product.ID]*product.Product{}
	// We need to keep track of the order because we are storing
	// products in a map. The order would change every time
	// if we loop over the map.
	order := []product.ID{}

	for rows.Next() {
		var (
			p             product.Product
			variationID   sql.NullInt64
			variationName sql.NullString
			quantity      sql.NullInt64
			brandID       sql.NullInt64
			brandName     sql.NullString
		)

		err := rows.Scan(
			&p.ID, &p.CreatedAt, &p.Name, &p.Price, &variationID, &variationName, &quantity, &brandID, &brandName,
		)
		if err != nil {
			return []product.Product{}, err
		}

		found, ok := pp[p.ID]
		if ok { // Second result of the same product. It means it is a variation
			v := product.Variation{
				ID:       product.VariationID(variationID.Int64),
				Name:     variationName.String,
				Quantity: int(quantity.Int64),
			}
			found.Variations = append(found.Variations, v)
			continue
		}

		// First time seeing the product. Brand can be null
		if brandID.Valid {
			p.Brand = &product.Brand{ID: product.BrandID(brandID.Int64), Name: brandName.String}
		}

		// If variation exists add it to the slice
		if variationID.Valid {
			v := product.Variation{
				ID:       product.VariationID(variationID.Int64),
				Name:     variationName.String,
				Quantity: int(quantity.Int64),
			}
			p.Variations = append(p.Variations, v)
		}
		pp[p.ID] = &p
		order = append(order, p.ID)
	}

	res := make([]product.Product, 0, len(order))
	for _, id := range order {
		res = append(res, *pp[id])
	}

	return res, nil
}

func (store ProductStore) InsertProduct(ctx context.Context, p product.Product) (product.Product, error) {
	var brandId sql.NullInt64
	if p.Brand != nil {
		brandId = sql.NullInt64{Int64: int64(p.Brand.ID), Valid: true}
	}
	err := store.db.QueryRowContext(ctx, `
		INSERT INTO products(created_at, name, price, brand_id) 
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, p.CreatedAt, p.Name, p.Price, brandId).Scan(&p.ID)
	if err != nil {
		return product.Product{}, fmt.Errorf("failed to insert product. error: %v", err)
	}

	return p, nil
}

func (store ProductStore) InsertProductVariation(ctx context.Context, ID product.ID, v product.Variation) (product.Variation, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return product.Variation{}, fmt.Errorf("failed to start transaction. %v", err)
	}

	err = tx.QueryRowContext(ctx, `INSERT INTO variations(name) VALUES($1) RETURNING id`, v.Name).Scan(&v.ID)
	if err != nil {
		msg := fmt.Sprintf("failed to insert variation. error: %v.", err)
		if err := tx.Rollback(); err != nil {
			msg += fmt.Sprintf(" transaction rollback error. %v", err)
		}
		return product.Variation{}, errors.New(msg)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO product_variations(quantity, product_id, variation_id) 
		VALUES ($1, $2, $3)
	`, v.Quantity, ID, v.ID)
	if err != nil {
		msg := fmt.Sprintf("failed to insert variation. error: %v.", err)
		if err := tx.Rollback(); err != nil {
			msg += fmt.Sprintf(" transaction rollback error. %v", err)
		}
		return product.Variation{}, errors.New(msg)
	}

	if err := tx.Commit(); err != nil {
		return product.Variation{}, fmt.Errorf("failed to commit transaction. %v", err)
	}

	return v, nil
}

func (store ProductStore) DeleteProduct(ctx context.Context, ID product.ID) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction. %v", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM product_variations WHERE product_id = $1`, ID); err != nil {
		tx.Rollback()
		msg := fmt.Sprintf("failed to delete product. error: %v.", err)
		if err := tx.Rollback(); err != nil {
			msg += fmt.Sprintf(" transaction rollback error. %v", err)
		}
		return errors.New(msg)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, ID); err != nil {
		tx.Rollback()
		msg := fmt.Sprintf("failed to delete product. error: %v.", err)
		if err := tx.Rollback(); err != nil {
			msg += fmt.Sprintf(" transaction rollback error. %v", err)
		}
		return errors.New(msg)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction. %v", err)
	}

	return nil
}

// BulkInsertProducts inserts product metadata in bulk. Without inserting brands and variations.
// Referenced brands has to be created first.
func (store ProductStore) BulkInsertProducts(ctx context.Context, pp []product.Product) ([]product.Product, error) {
	placeHolders := make([]string, 0, len(pp)) // ($1, $2, $3, $4), ($5, $6, $7, $8) ...
	values := make([]interface{}, 0, len(pp))

	for i, p := range pp { // Generate place holders for each product
		ph := make([]string, 0, 4) // Create a place holder slice. For 4 columns
		for j := 1; j < 5; j++ {   // Add placeholder 4 times
			ph = append(ph, fmt.Sprintf("$%d", i*4+j))
		}
		// Append the joined placeholder
		placeHolders = append(placeHolders, fmt.Sprintf("(%s)", strings.Join(ph, ",")))

		var brandID sql.NullInt64
		if p.Brand != nil {
			brandID = sql.NullInt64{Int64: int64(p.Brand.ID), Valid: true}
		}
		values = append(values, p.CreatedAt, p.Name, p.Price, brandID)
	}

	stmt := fmt.Sprintf(`
		WITH inserted AS (
			INSERT INTO products(created_at, name, price, brand_id)
			VALUES %s 
			RETURNING *
		)
		SELECT p.id, p.created_at, p.name, p.price, b.id, b.name
		FROM inserted p
		LEFT JOIN brands b ON b.id = p.brand_id
	`, strings.Join(placeHolders, ","))

	rows, err := store.db.QueryContext(ctx, stmt, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk insert products. error: %v", err)
	}
	defer rows.Close()

	inserted := make([]product.Product, 0, len(pp))

	for rows.Next() {
		var (
			p         product.Product
			brandID   sql.NullInt64
			brandName sql.NullString
		)
		if err := rows.Scan(&p.ID, &p.CreatedAt, &p.Name, &p.Price, &brandID, &brandName); err != nil {
			return nil, fmt.Errorf("failed to bulk insert products. error: %v", err)
		}

		if brandID.Valid {
			p.Brand = &product.Brand{ID: product.BrandID(brandID.Int64), Name: brandName.String}
		}

		inserted = append(inserted, p)
	}

	return inserted, nil
}

func (store ProductStore) SearchProducts(ctx context.Context, query string) ([]product.Product, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, created_at, name, price 
		FROM products
		WHERE name ILIKE CONCAT($1::text, '%')
		LIMIT 20
	`, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search products. error: %v", err)
	}
	defer rows.Close()

	pp := []product.Product{}
	for rows.Next() {
		var p product.Product
		if err := rows.Scan(&p.ID, &p.CreatedAt, &p.Name, &p.Price); err != nil {
			return nil, fmt.Errorf("failed to search products. error: %v", err)
		}
		pp = append(pp, p)
	}

	return pp, nil
}

func (store ProductStore) GetProductsOffsetPagination(ctx context.Context, page, rowsPerPage *int) ([]product.Product, int, error) {
	var pagination string
	values := make([]interface{}, 0, 2)
	if page != nil {
		rpp := 25
		if rowsPerPage != nil {
			rpp = *rowsPerPage
		}
		offset := (*page - 1) * rpp
		pagination = "LIMIT $1 OFFSET $2"
		values = append(values, rpp, offset)
	}
	stmt := fmt.Sprintf(`
		SELECT p.id, created_at, p.name, price, v.id, v.name, pv.quantity, b.id, b.name,
		(SELECT COUNT(*) FROM products) AS total
		FROM (SELECT * FROM products p %s) p
		LEFT JOIN product_variations pv ON p.id = pv.product_id
		LEFT JOIN variations v ON pv.variation_id = v.id
		LEFT JOIN brands b ON p.brand_id = b.id
		ORDER BY created_at DESC
	`, pagination)
	rows, err := store.db.Query(stmt, values...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get products. error: %v", err)
	}
	defer rows.Close()

	var total int
	pp := map[product.ID]*product.Product{}
	order := []product.ID{}

	for rows.Next() {
		var (
			p             product.Product
			variationID   sql.NullInt64
			variationName sql.NullString
			quantity      sql.NullInt64
			brandID       sql.NullInt64
			brandName     sql.NullString
		)

		err := rows.Scan(
			&p.ID, &p.CreatedAt, &p.Name, &p.Price, &variationID,
			&variationName, &quantity, &brandID, &brandName, &total,
		)
		if err != nil {
			return []product.Product{}, 0, err
		}

		found, ok := pp[p.ID]
		if ok { // Second result of the same product. It means it is a variation
			v := product.Variation{
				ID:       product.VariationID(variationID.Int64),
				Name:     variationName.String,
				Quantity: int(quantity.Int64),
			}
			found.Variations = append(found.Variations, v)
			continue
		}

		// First time seeing the product. Brand can be null
		if brandID.Valid {
			p.Brand = &product.Brand{ID: product.BrandID(brandID.Int64), Name: brandName.String}
		}

		// If variation exists add it to the slice
		if variationID.Valid {
			v := product.Variation{
				ID:       product.VariationID(variationID.Int64),
				Name:     variationName.String,
				Quantity: int(quantity.Int64),
			}
			p.Variations = append(p.Variations, v)
		}
		pp[p.ID] = &p
		order = append(order, p.ID)
	}

	res := make([]product.Product, 0, len(order))
	for _, id := range order {
		res = append(res, *pp[id])
	}

	return res, total, nil
}

type OrderBy int

const (
	OrderByCreatedAt OrderBy = iota
	OrderByPrice
)

type OrderByDirection int

const (
	OrderByDirectionDESC OrderByDirection = iota
	OrderByDirectionASC
)

type Params struct {
	StartPrice       *int
	EndPrice         *int
	BrandIDs         []product.BrandID
	OrderBy          OrderBy
	OrderByDirection OrderByDirection
	Page             *int
	Limit            *int
}

func (store ProductStore) GetProductsFilteredSorted(ctx context.Context, params Params) ([]product.Product, int, error) {
	extraStmts := []string{}
	values := []interface{}{}
	var (
		orderBy          string
		orderByDirection string
		pagination       string
	)

	if len(params.BrandIDs) != 0 {
		placeholders := make([]string, 0, len(params.BrandIDs))
		for _, id := range params.BrandIDs {
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(values)+1))
			values = append(values, id)
		}
		extraStmts = append(extraStmts, fmt.Sprintf("p.brand_id IN (%s)", strings.Join(placeholders, ",")))
	}

	if params.StartPrice != nil {
		extraStmts = append(extraStmts, fmt.Sprintf("p.price >= $%d", len(values)+1))
		values = append(values, *params.StartPrice)
	}

	if params.EndPrice != nil {
		extraStmts = append(extraStmts, fmt.Sprintf("p.price <= $%d", len(values)+1))
		values = append(values, *params.EndPrice)
	}

	if params.Page != nil {
		rpp := 25
		if params.Limit != nil {
			rpp = *params.Limit
		}
		offset := (*params.Page - 1) * rpp
		pagination = fmt.Sprintf("LIMIT $%d OFFSET $%d", len(values)+1, len(values)+2)
		values = append(values, rpp, offset)
	}

	var filters string
	totalSubQuery := `SELECT COUNT(*) FROM products AS p`
	if len(extraStmts) > 0 {
		where := fmt.Sprintf(" WHERE %s", strings.Join(extraStmts, " AND "))
		filters += where
		totalSubQuery += where
	}

	switch params.OrderBy {
	case OrderByPrice:
		orderBy = "ORDER BY p.price"
	default:
		orderBy = "ORDER BY p.created_at"
	}

	switch params.OrderByDirection {
	case OrderByDirectionASC:
		orderByDirection = "ASC"
	default:
		orderByDirection = "DESC"
	}

	stmt := fmt.Sprintf(`
		SELECT p.id, created_at, p.name, price, v.id, 
		v.name, pv.quantity, b.id, b.name, (%s) AS total
		FROM (
			SELECT * FROM products AS p %s 
			%s %s
			%s
		) AS p
		LEFT JOIN product_variations AS pv ON p.id = pv.product_id
		LEFT JOIN variations AS v ON pv.variation_id = v.id
		LEFT JOIN brands AS b ON p.brand_id = b.id
		%s %s, quantity
	`, totalSubQuery, filters, orderBy, orderByDirection, pagination, orderBy, orderByDirection)

	rows, err := store.db.Query(stmt, values...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get products. error: %v", err)
	}
	defer rows.Close()

	var total int
	pp := map[product.ID]*product.Product{}
	order := []product.ID{}

	for rows.Next() {
		var (
			p             product.Product
			variationID   sql.NullInt64
			variationName sql.NullString
			quantity      sql.NullInt64
			brandID       sql.NullInt64
			brandName     sql.NullString
		)

		err := rows.Scan(
			&p.ID, &p.CreatedAt, &p.Name, &p.Price, &variationID,
			&variationName, &quantity, &brandID, &brandName, &total,
		)
		if err != nil {
			return []product.Product{}, 0, err
		}

		found, ok := pp[p.ID]
		if ok { // Second result of the same product. It means it is a variation
			v := product.Variation{
				ID:       product.VariationID(variationID.Int64),
				Name:     variationName.String,
				Quantity: int(quantity.Int64),
			}
			found.Variations = append(found.Variations, v)
			continue
		}

		// First time seeing the product. Brand can be null
		if brandID.Valid {
			p.Brand = &product.Brand{ID: product.BrandID(brandID.Int64), Name: brandName.String}
		}

		// If variation exists add it to the slice
		if variationID.Valid {
			v := product.Variation{
				ID:       product.VariationID(variationID.Int64),
				Name:     variationName.String,
				Quantity: int(quantity.Int64),
			}
			p.Variations = append(p.Variations, v)
		}
		pp[p.ID] = &p
		order = append(order, p.ID)
	}

	res := make([]product.Product, 0, len(order))
	for _, id := range order {
		res = append(res, *pp[id])
	}

	return res, total, nil
}
