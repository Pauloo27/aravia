package impl

import (
	"github.com/Pauloo27/aravia"
	"github.com/gofiber/fiber/v2"
)

type FiberServer struct {
	Fiber *fiber.App
}

func NewFiberServer(server *fiber.App) FiberServer {
	return FiberServer{
		Fiber: server,
	}
}

func (s FiberServer) Listen(bindAddr string) error {
	return s.Fiber.Listen(bindAddr)
}

func (s FiberServer) Route(method aravia.HttpMethod, path string, handler aravia.Handler) {
	s.Fiber.Add(string(method), path, func(ctx *fiber.Ctx) error {
		var params = make(map[string]string)
		var query = make(map[string]string)
		for _, param := range ctx.Route().Params {
			params[param] = ctx.Params(param)
		}
		ctx.Context().QueryArgs().VisitAll(func(key, value []byte) {
			query[string(key)] = string(value)
		})

		response := handler(aravia.Request{
			Body:    ctx.Body(),
			Headers: ctx.GetReqHeaders(),
			Path:    ctx.Path(),
			Method:  aravia.HttpMethod(ctx.Method()),
			Params:  params,
			Query:   query,
		})
		if response.Data == nil {
			return ctx.Status(int(response.StatusCode)).Send([]byte{})
		}
		return ctx.Status(int(response.StatusCode)).JSON(response.Data)
	})
}
