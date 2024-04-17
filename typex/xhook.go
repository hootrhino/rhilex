package typex

// XHook for enhancement rhilex with golang
type XHook interface {
	Work(data string) error
	Error(error)
	Name() string
}
