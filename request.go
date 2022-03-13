package aravia

type Request struct {
	Body    []byte
	Headers map[string]string
	Path    string
}
