package impl

import (
	"github.com/Pauloo27/aravia"
	"github.com/gofiber/fiber/v2"
)

type FiberServer struct {
	app *fiber.App
}

func NewFiberServer() FiberServer {
	return FiberServer{
		app: fiber.New(),
	}
}

func (s FiberServer) Listen(bindAddr string) error {
	return s.app.Listen(bindAddr)
}

func (s FiberServer) Route(method aravia.HttpMethod, path string, handler aravia.Handler) {
	s.app.Add(string(method), path, func(ctx *fiber.Ctx) error {
		response := handler(aravia.Request{
			Body:    ctx.Body(),
			Headers: ctx.GetReqHeaders(),
			Path:    ctx.Path(),
			Method:  aravia.HttpMethod(ctx.Method()),
		})
		return ctx.Status(int(response.StatusCode)).JSON(response.Data)
	})
}
