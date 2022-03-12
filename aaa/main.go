package main

import (
	"github.com/Pauloo27/aravia"
	"github.com/Pauloo27/aravia/impl"
	"github.com/Pauloo27/logger"
)

type UserController struct {
}

type User struct {
	Name  string
	Email string
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

func main() {
	app := aravia.App{
		Server: impl.NewFiberServer(),
	}
	app.RegisterController(UserController{})
	app.Listen(":8023")
}
