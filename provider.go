package configuration

type Provider interface {
	GetBuffer() (*TreeBuffer, error)
}
