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
	tStatus      = reflect.TypeOf(HttpStatus(200))
	tRequest     = reflect.TypeOf(Request{})
	reBodyInput  = regexp.MustCompile(`^\w+BodyInput$`)
	reQueryInput = regexp.MustCompile(`^\w+QueryInput$`)
)

type InputType string

const (
	InputBody    = InputType("body")
	InputRequest = InputType("request")
	InputQuery   = InputType("query")
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
		if reQueryInput.MatchString(in.Name()) {
			inputs[i] = InputQuery
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
	for i := 0; i < num; i++ {
		item := f(i)
		if item.AssignableTo(targetType) {
			return i
		}
	}
	return -1
}

func parseAndValidate(tInput reflect.Type, data []byte) (interface{}, error) {
	input := reflect.New(tInput).Interface()
	err := json.Unmarshal(data, input)
	if err != nil {
		return nil, HttpError{
			StatusCode: StatusUnprocessableEntity,
			Message:    err.Error(),
			Data:       Map{"error": err.Error()},
		}
	}
	errs := Validate(input)
	if len(errs) != 0 {
		return nil, HttpError{
			StatusCode: StatusBadRequest,
			Message:    "validation error",
			Data:       Map{"error": "validation error", "more": errs},
		}
	}
	return input, nil
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
	info.Path = strings.Trim(info.Path, "/")

	cMethods := cType.NumMethod()

	routeByHandlerName := make(map[string]string)
	routeMethodByHandlerName := make(map[string]HttpMethod)
	for route, routeInfo := range info.Routes {
		routeByHandlerName[routeInfo.HandlerName] = route
		routeMethodByHandlerName[routeInfo.HandlerName] = routeInfo.Method
	}

	logger.Info("[CONTROLLER]", info.Path, cMethods)
	for i := 0; i < cMethods; i++ {
		m := cType.Method(i)
		mName := m.Name
		// skip specials or private functions
		if mName == "Init" || !m.IsExported() {
			continue
		}

		var (
			path   string
			method HttpMethod
		)

		mappedPath, found := routeByHandlerName[mName]
		method = routeMethodByHandlerName[mName]

		path = "/" + info.Path
		if found {
			path = path + "/" + strings.Trim(mappedPath, "/")
		} else {
			words := GetWords(mName)
			if len(words) == 0 {
				return errors.New("invalid handler name")
			}
			method = HttpMethod(strings.ToUpper(words[0]))
			sb := strings.Builder{}
			sb.WriteString(path)

			for _, word := range words[1:] {
				if strings.HasPrefix(word, "_") {
					word = ":" + word[1:]
				}
				sb.WriteString("/")
				sb.WriteString(strings.ToLower(word))
			}
			path = sb.String()
		}

		logger.Info("[ROUTE]", method, path)

		statusOutIdx := findType(m.Type.NumOut(), m.Type.Out, tStatus)
		logger.Debug("idx", statusOutIdx)

		inputs := listInputs(m)

		outLen := m.Type.NumOut()

		callable := cValue.Method(i)

		a.Route(HttpMethod(method), path, func(req Request) Response {
			var params = make([]reflect.Value, len(inputs))
			for i, in := range inputs {
				switch in {
				case InputRequest:
					params[i] = reflect.ValueOf(req)
				case InputQuery:
					// im really sorry mom
					encodedQuery, _ := json.Marshal(req.Query)
					query, err := parseAndValidate(m.Type.In(i+1), encodedQuery)
					if err != nil {
						httpError := err.(HttpError)
						return Response{
							StatusCode: httpError.StatusCode,
							Data:       httpError.Data,
						}
					}
					params[i] = reflect.ValueOf(query).Elem()
				case InputBody:
					body, err := parseAndValidate(m.Type.In(i+1), req.Body)
					if err != nil {
						httpError := err.(HttpError)
						return Response{
							StatusCode: httpError.StatusCode,
							Data:       httpError.Data,
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
			var data interface{}

			if outLen == 0 || statusOutIdx == 0 {
				status = StatusNoContent
				data = nil
			} else {
				data = out[0].Interface()
			}

			if statusOutIdx != -1 {
				status = out[statusOutIdx].Interface().(HttpStatus)
			}
			return Response{
				StatusCode: status,
				Data:       data,
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
