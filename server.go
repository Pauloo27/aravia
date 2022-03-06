package aravia

type Server interface {
	Listen(bindAddr string) error
	Route(method HttpMethod, path string, handler Handler)
}
