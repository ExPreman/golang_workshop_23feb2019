package handler

import (
	"net/http"

	middleware "github.com/payfazz/go-middleware"
	"github.com/payfazz/go-middleware/common/kv"
	"github.com/payfazz/go-middleware/common/logger"
	"github.com/payfazz/go-middleware/common/paniclogger"
	"github.com/payfazz/go-router/path"

	"github.com/win-t/golang_workshop_23feb2019/internal/http/middleware/cors"
)

func Compile(app *App) http.HandlerFunc {
	return middleware.Compile(
		logger.New(
			logger.DefaultLogger(app.InfoLog),
		),

		paniclogger.New(
			15,
			paniclogger.DefaultLogger(app.ErrLog),
		),

		kv.New(),

		cors.OriginAll(),

		path.H{
			"/":        app.hello(),
			"/counter": app.counter(),
			"/reset":   app.reset(),
		}.C(),
	)
}
