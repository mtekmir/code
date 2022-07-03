package product

import "time"

type VariationID int64

// Represents a variation of a product. XS, S, L, red, blue etc.
type Variation struct {
	ID       VariationID
	Name     string
	Quantity int
}

type BrandID int64

// Represents a brand of a product.
type Brand struct {
	ID   BrandID
	Name string
}

type ID int64

// Represents a product
type Product struct {
	ID         ID
	CreatedAt  time.Time
	Name       string
	Price      int         // Price in cents
	Brand      *Brand      `json:",omitempty"`
	Variations []Variation `json:",omitempty"`
}
