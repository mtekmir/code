package product

type VariationID int64

// Represents a variation of a product. XS, S, L, red, blue etc.
type Variation struct {
	ID   VariationID
	Name string
}

type ID int64

// Represents a product
type Product struct {
	ID        ID
	Barcode   string
	Name      string
	Price     string
	Quantity  int
	Variation *Variation
}
