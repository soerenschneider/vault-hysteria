package messagefilters

type MessageFilter interface {
	Accept(string) bool
}
