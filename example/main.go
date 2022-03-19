package main

import (
	"github.com/Pauloo27/aravia"
	"github.com/Pauloo27/aravia/impl"
	"github.com/Pauloo27/logger"
)

type UserController struct {
}

type User struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"e-mail" validate:"required,email"`
}

type FilterQueryInput struct {
	Type  string `json:"type" validate:"required"`
	Email string `json:"e-mail,omitempty" validate:"omitempty,email"`
}

type UserBodyInput User

func (c UserController) Init() *aravia.ControllerInfo {
	return &aravia.ControllerInfo{
		Path: "users",
		Routes: map[string]aravia.RouteInfo{
			":id": {
				Method:      "Get",
				HandlerName: "GetById",
			},
		},
	}
}

func (UserController) Post(body UserBodyInput) (string, aravia.HttpStatus) {
	logger.Success("called =)", body)
	return "post =)", 418
}

func (UserController) Delete() {
	logger.Info("delete was called =)")
}

func (UserController) DeleteTest() aravia.HttpStatus {
	return aravia.StatusRequestEntityTooLarge
}

func (UserController) Get(filters FilterQueryInput) string {
	logger.Success("filters", filters)
	return "get =)"
}

func (UserController) GetById(req aravia.Request) string {
	return "get by id " + req.Params["id"]
}

type AuthMiddleware struct {
}

func (AuthMiddleware) Run(req aravia.Request) (res *aravia.Response) {
	logger.Debug("[MIDDLEWARE] ", req.Path)
	if _, found := req.Headers["Authorization"]; found {
		return nil
	}
	return &aravia.Response{
		StatusCode: aravia.StatusUnauthorized,
		Data: aravia.Map{
			"error": "missing Authorization header",
		},
	}
}

func main() {
	app := aravia.App{
		Server: impl.NewFiberServer(),
	}
	app.Use(AuthMiddleware{})
	app.With(UserController{})
	app.Listen(":8023")
}
