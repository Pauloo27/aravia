package aravia

type RouteInfo struct {
	Method      HttpMethod
	HandlerName string
}

type ControllerInfo struct {
	Path   string
	Routes map[string]RouteInfo
}

type Controller interface {
	Init() *ControllerInfo
}
