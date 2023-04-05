package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"code.com/product"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return Store{db: db}
}

type Cursors struct {
	Prev string
	Next string
}

func (store Store) GetProducts(
	ctx context.Context, cursors Cursors, limit int,
) ([]product.Product, Cursors, error) {
	if limit == 0 {
		return []product.Product{}, Cursors{}, errors.New("limit cannot be zero")
	}
	if cursors.Next != "" && cursors.Prev != "" {
		return []product.Product{}, Cursors{}, errors.New("two cursors cannot be provided at the same time")
	}

	values := make([]interface{}, 0, 4)
	rowsLeftQuery := "SELECT COUNT(*) FROM products p"
	pagination := ""

	if cursors.Next != "" {
		f := fmt.Sprintf("p.created_at < $%d", len(values)+1)
		rowsLeftQuery += fmt.Sprintf(" WHERE %s", f)
		pagination += fmt.Sprintf("WHERE %s ORDER BY created_at DESC LIMIT $%d", f, len(values)+2)
		values = append(values, cursors.Next, limit)
	}

	if cursors.Prev != "" {
		f := fmt.Sprintf("p.created_at > $%d", len(values)+1)
		rowsLeftQuery += fmt.Sprintf(" WHERE %s", f)
		pagination += fmt.Sprintf("WHERE %s ORDER BY created_at ASC LIMIT $%d", f, len(values)+2)
		values = append(values, cursors.Prev, limit)
	}

	if cursors.Next == "" && cursors.Prev == "" {
		pagination = fmt.Sprintf(" ORDER BY p.created_at DESC LIMIT $%d", len(values)+1)
		values = append(values, limit)
	}

	stmt := fmt.Sprintf(`
		WITH p AS (
			SELECT * FROM products p %s
		)
		SELECT id, created_at, name, 
		(%s) AS rows_left,
		(SELECT COUNT(*) FROM products) AS total
		FROM p
		ORDER BY created_at DESC
	`, pagination, rowsLeftQuery)

	rows, err := store.db.Query(stmt, values...)
	if err != nil {
		return nil, Cursors{}, fmt.Errorf("failed to get products. error: %v", err)
	}
	defer rows.Close()

	var (
		rowsLeft int
		total    int
		pp       = []product.Product{}
	)

	for rows.Next() {
		var p product.Product

		err := rows.Scan(&p.ID, &p.CreatedAt, &p.Name, &rowsLeft, &total)
		if err != nil {
			return []product.Product{}, Cursors{}, err
		}

		pp = append(pp, p)
	}

	var (
		prevCursor string // cursor we return when there is a prev page
		nextCursor string // cursor we return when there is a next page
	)

	//  A     B     C			D			E
	//  |-----|-----|-----|-----|

	// When we receive a next cursor, direction = A->E
	// When we receive a prev cursor, direction = E->A

	switch {

	// *If there are no results we don't have to compute the cursors
	case rowsLeft < 0:

	// *On A, direction A->E (going forward), return only next cursor
	case cursors.Prev == "" && cursors.Next == "":
		nextCursor = pp[len(pp)-1].CreatedAt.UTC().Format(time.RFC3339)

	// *On E, direction A->E (going forward), return only prev cursor
	case cursors.Next != "" && rowsLeft == len(pp):
		prevCursor = pp[0].CreatedAt.UTC().Format(time.RFC3339)

	// *On A, direction E->A (going backwards), return only next cursor
	case cursors.Prev != "" && rowsLeft == len(pp):
		nextCursor = pp[len(pp)-1].CreatedAt.UTC().Format(time.RFC3339)

	// *On E, direction E->A (going backwards), return only prev cursor
	case cursors.Prev != "" && total == rowsLeft:
		prevCursor = pp[0].CreatedAt.UTC().Format(time.RFC3339)

	// *Somewhere in the middle
	default:
		nextCursor = pp[len(pp)-1].CreatedAt.UTC().Format(time.RFC3339)
		prevCursor = pp[0].CreatedAt.UTC().Format(time.RFC3339)

	}

	return pp, Cursors{Prev: prevCursor, Next: nextCursor}, nil
}
