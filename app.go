package aravia

import (
	"errors"
	"reflect"
	"strings"

	"github.com/Pauloo27/logger"
)

type App struct {
	Controllers []Controller
	Server      Server
}

var (
	tStatus = reflect.TypeOf(HttpStatus(200))
)

func (a *App) routeController(c Controller) error {
	cType := reflect.TypeOf(c)
	cValue := reflect.ValueOf(c)

	cName := cType.Name()
	cMethods := cType.NumMethod()
	cRoot := strings.ToLower(GetWords(cName)[0])

	logger.Info("[CONTROLLER]", cName, cMethods)
	for i := 0; i < cMethods; i++ {
		m := cType.Method(i)
		mName := m.Name
		words := GetWords(mName)
		if len(words) == 0 {
			return errors.New("invalid handler name")
		}
		method := strings.ToUpper(words[0])
		sb := strings.Builder{}
		sb.WriteString("/")
		sb.WriteString(cRoot)

		for _, word := range words[1:] {
			sb.WriteString("/")
			sb.WriteString(strings.ToLower(word))
		}

		path := sb.String()

		logger.Infof("[ROUTE] %s %s", method, path)

		statusOutIdx := -1

		for outIdx := 1; outIdx < m.Type.NumOut(); outIdx++ {
			out := m.Type.Out(outIdx)
			if out.AssignableTo(tStatus) {
				statusOutIdx = outIdx
				break
			}
		}

		callable := cValue.Method(i)

		a.Route(HttpMethod(method), path, func(req Request) Response {
			out := callable.Call([]reflect.Value{})
			status := StatusOK
			if statusOutIdx != -1 {
				status = out[statusOutIdx].Interface().(HttpStatus)
			}
			return Response{
				StatusCode: status,
				Data:       out[0].Interface(),
			}
		})
	}
	return nil
}

func (a *App) routeControllers() error {
	for _, c := range a.Controllers {
		if err := a.routeController(c); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) RegisterController(controller Controller) {
	a.Controllers = append(a.Controllers, controller)
}

func (a *App) Listen(bindAddr string) error {
	if err := a.routeControllers(); err != nil {
		return err
	}
	return a.Server.Listen(bindAddr)
}

func (a *App) Route(method HttpMethod, path string, handler Handler) {
	a.Server.Route(method, path, handler)
}
