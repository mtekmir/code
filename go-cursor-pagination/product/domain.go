package product

import "time"

type ID int

type Product struct {
	ID        ID
	CreatedAt time.Time
	Name      string
}
