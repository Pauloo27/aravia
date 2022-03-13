package aravia

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"

	"github.com/Pauloo27/logger"
)

type App struct {
	Middlewares []Middleware
	Controllers []Controller
	Server      Server
}

var (
	tStatus     = reflect.TypeOf(HttpStatus(200))
	tRequest    = reflect.TypeOf(Request{})
	reBodyInput = regexp.MustCompile(`^\w+BodyInput$`)
)

type InputType string

const (
	InputBody    = InputType("body")
	InputRequest = InputType("request")
	InputInvalid = InputType("")
)

func listInputs(method reflect.Method) []InputType {
	inCount := method.Type.NumIn() - 1
	if inCount == 0 {
		return nil
	}
	var inputs = make([]InputType, inCount)
	for i := 0; i < inCount; i++ {
		in := method.Type.In(i + 1)
		if reBodyInput.MatchString(in.Name()) {
			inputs[i] = InputBody
			continue
		}
		if in.AssignableTo(tRequest) {
			inputs[i] = InputRequest
			continue
		}
		logger.Fatal("invalid handler input type", in.Name())
	}
	return inputs
}

func findByNameRe(num int, f func(i int) reflect.Type, re *regexp.Regexp) int {
	for i := 1; i < num; i++ {
		item := f(i)
		if re.MatchString(item.Name()) {
			return i
		}
	}
	return -1
}

func findType(num int, f func(i int) reflect.Type, targetType reflect.Type) int {
	for i := 1; i < num; i++ {
		item := f(i)
		if item.AssignableTo(targetType) {
			return i
		}
	}
	return -1
}

func (a *App) routeController(c Controller) error {
	cType := reflect.TypeOf(c)
	cValue := reflect.ValueOf(c)

	info := c.Init()
	if info == nil {
		info = &ControllerInfo{
			Path: strings.ToLower(GetWords(cType.Name())[0]),
		}
	}

	cMethods := cType.NumMethod()

	logger.Info("[CONTROLLER]", info.Path, cMethods)
	for i := 0; i < cMethods; i++ {
		m := cType.Method(i)
		mName := m.Name
		// skip specials or private functions
		if mName == "Init" || !m.IsExported() {
			continue
		}
		words := GetWords(mName)
		if len(words) == 0 {
			return errors.New("invalid handler name")
		}
		method := strings.ToUpper(words[0])
		sb := strings.Builder{}
		sb.WriteString("/")
		sb.WriteString(info.Path)

		for _, word := range words[1:] {
			if strings.HasPrefix(word, "_") {
				word = ":" + word[1:]
			}
			sb.WriteString("/")
			sb.WriteString(strings.ToLower(word))
		}

		path := sb.String()

		logger.Info("[ROUTE]", method, path)

		statusOutIdx := findType(m.Type.NumOut(), m.Type.Out, tStatus)

		inputs := listInputs(m)

		callable := cValue.Method(i)

		a.Route(HttpMethod(method), path, func(req Request) Response {
			var params = make([]reflect.Value, len(inputs))
			for i, in := range inputs {
				switch in {
				case InputRequest:
					params[i] = reflect.ValueOf(req)
				case InputBody:
					bodyIn := m.Type.In(i + 1)
					body := reflect.New(bodyIn).Interface()
					err := json.Unmarshal(req.Body, body)
					if err != nil {
						return Response{
							Data: Map{
								"error": "invalid body input",
							},
							StatusCode: StatusUnprocessableEntity,
						}
					}
					errs := Validate(body)
					if len(errs) != 0 {
						return Response{
							Data: Map{
								"error": "validation error",
								"more":  errs,
							},
							StatusCode: StatusBadRequest,
						}
					}
					params[i] = reflect.ValueOf(body).Elem()
				}
			}

			for _, middleware := range a.Middlewares {
				res := middleware.Run(req)
				if res != nil {
					return *res
				}
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

func (a *App) With(controller Controller) {
	a.Controllers = append(a.Controllers, controller)
}

func (a *App) Use(middleware Middleware) {
	a.Middlewares = append(a.Middlewares, middleware)
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
