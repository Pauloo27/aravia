package aravia

type RouteInfo struct {
	Path   string
	Method HttpMethod
}

type ControllerInfo struct {
	Path   string
	Routes map[string]RouteInfo
}

type Controller interface {
	Init() *ControllerInfo
}
