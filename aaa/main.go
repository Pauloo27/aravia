package main

import (
	"github.com/Pauloo27/aravia"
	"github.com/Pauloo27/aravia/impl"
	"github.com/Pauloo27/logger"
)

type UserController struct {
}

func (UserController) Post() (string, aravia.HttpStatus) {
	logger.Success("called =)")
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
