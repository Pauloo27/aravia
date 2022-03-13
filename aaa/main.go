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

type UserBodyInput User

func (UserController) Post(body UserBodyInput) (string, aravia.HttpStatus) {
	logger.Success("called =)", body)
	return "post =)", 418
}

func (UserController) Get() string {
	return "get =)"
}

func (UserController) GetName() string {
	return "get name =)"
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
