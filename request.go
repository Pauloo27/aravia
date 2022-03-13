package aravia

type Request struct {
	Method  HttpMethod
	Body    []byte
	Headers map[string]string
	Path    string
}
