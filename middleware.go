package aravia

type Middleware interface {
	Run(Request) *Response
}
