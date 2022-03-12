package aravia

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"

	"github.com/Pauloo27/logger"
	"github.com/gofiber/fiber/v2"
)

type App struct {
	Controllers []Controller
	Server      Server
}

var (
	tStatus     = reflect.TypeOf(HttpStatus(200))
	reBodyInput = regexp.MustCompile(`^\w+BodyInput$`)
)

func findByNameRe(num int, f func(i int) reflect.Type, re *regexp.Regexp) int {
	for i := 1; i < num; i++ {
		out := f(i)
		if re.MatchString(out.Name()) {
			return i
		}
	}
	return -1
}

func findType(num int, f func(i int) reflect.Type, targetType reflect.Type) int {
	for i := 1; i < num; i++ {
		out := f(i)
		if out.AssignableTo(targetType) {
			return i
		}
	}
	return -1
}

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

		logger.Info("[ROUTE]", method, path)

		statusOutIdx := findType(m.Type.NumOut(), m.Type.Out, tStatus)

		bodyInIdx := findByNameRe(m.Type.NumIn(), m.Type.In, reBodyInput)

		callable := cValue.Method(i)

		a.Route(HttpMethod(method), path, func(req Request) Response {
			var params []reflect.Value
			if bodyInIdx != -1 {
				bodyIn := m.Type.In(bodyInIdx)
				body := reflect.New(bodyIn).Interface()
				err := json.Unmarshal(req.Body, body)
				if err != nil {
					return Response{
						Data: fiber.Map{
							"error": "invalid body input",
						},
						StatusCode: StatusUnprocessableEntity,
					}
				}
				params = append(params,
					reflect.ValueOf(body).Elem(),
				)
			}

			out := callable.Call(params)
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
