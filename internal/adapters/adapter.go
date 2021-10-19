package adapters

type Adapter interface {
	Start(chan string) error
	Stop() error
	Name() string
}
