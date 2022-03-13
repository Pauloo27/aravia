package aravia

type Request struct {
	Method  HttpMethod
	Body    []byte
	Headers map[string]string
	Path    string
	Params  map[string]string
	Query   map[string]string
}
