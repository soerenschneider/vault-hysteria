package any

const FilterName = "any"

type AnyFilter struct{}

func (c *AnyFilter) Accept(msg string) bool {
	return true
}
