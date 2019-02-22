package handler

import (
	"fmt"
	"log"
	"net/http"

	pfhandler "github.com/payfazz/go-handler"
	"github.com/payfazz/go-handler/defresponse"
	middleware "github.com/payfazz/go-middleware"
	"github.com/payfazz/go-router/method"
	"github.com/payfazz/go-router/segment"

	"github.com/win-t/golang_workshop_23feb2019/internal/http/middleware/auth"
	"github.com/win-t/golang_workshop_23feb2019/internal/storage"
)

type App struct {
	InfoLog          *log.Logger
	ErrLog           *log.Logger
	Storage          *storage.Storage
	ServerShutdownCh <-chan struct{}
}

func (a *App) hello() http.HandlerFunc {
	return middleware.Compile(
		segment.E,

		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprint(w, "Hello World")
		},
	)
}

func (a *App) counter() http.HandlerFunc {
	return middleware.Compile(
		segment.E,
		auth.New(a.ErrLog, a.Storage),

		method.H{
			"OPTIONS": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Headers", "Authorization")
			},

			"GET": pfhandler.Handler(func(r *http.Request) pfhandler.Response {
				name := auth.Name(r)
				if name == "" {
					return defresponse.Status(401)
				}

				ctx := r.Context()
				if err := a.Storage.IncCounter(ctx, name); err != nil {
					a.ErrLog.Println(err)
					return defresponse.Status(500)
				}

				counter, err := a.Storage.GetCounter(ctx, name)
				if err != nil {
					a.ErrLog.Println(err)
					return defresponse.Status(500)
				}

				worker := a.Storage.NewWorkerConnection()
				defer worker.Close()
				workID, err := worker.PostJob(ctx, counter)
				if err != nil {
					a.ErrLog.Println(err)
					return defresponse.Status(500)
				}

				var score int64

				workerNotifCh := worker.NotifCh()
				doneCh := ctx.Done()

			LOOP:
				for {
					select {
					case <-doneCh:
						a.ErrLog.Println("Context closed")
						return defresponse.Status(500)
					case <-workerNotifCh:
						score, err = worker.GetJobResult(ctx, workID)
						if err != nil {
							if err == storage.ErrNotFound {
								continue LOOP
							}
							a.ErrLog.Println(err)
							return defresponse.Status(500)
						}
						break LOOP
					case <-a.ServerShutdownCh:
						return defresponse.Status(503)
					}
				}

				return defresponse.JSON(200, struct {
					Counter int64 `json:"counter"`
					Score   int64 `josn:"score"`
				}{
					Counter: counter,
					Score:   score,
				})

			}).ServeHTTP,
		}.C(),
	)
}

func (a *App) reset() http.HandlerFunc {
	return middleware.Compile(
		segment.E,
		auth.New(a.ErrLog, a.Storage),

		method.H{
			"OPTIONS": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Headers", "Authorization")
			},

			"POST": pfhandler.Handler(func(r *http.Request) pfhandler.Response {
				isAdmin := auth.IsAdmin(r)
				if !isAdmin {
					return defresponse.Status(401)
				}

				if err := a.Storage.ResetCounter(r.Context()); err != nil {
					a.ErrLog.Println(err)
					return defresponse.Status(500)
				}

				return defresponse.Status(200)

			}).ServeHTTP,
		}.C(),
	)
}
