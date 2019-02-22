package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/win-t/golang_workshop_23feb2019/internal/config"
	"github.com/win-t/golang_workshop_23feb2019/internal/http/handler"
	"github.com/win-t/golang_workshop_23feb2019/internal/logger"
	"github.com/win-t/golang_workshop_23feb2019/internal/storage"
)

func main() {
	showEnv := flag.Bool("env", false, "Print config from env and exit")
	flag.Parse()

	if *showEnv {
		config.PrintAll()
		return
	}

	if len(flag.Args()) != 0 {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	infoLog, errLog := logger.Open(!config.DISABLE_PLAIN_LOG())

	s, err := storage.Open(config.PG_URL(), errLog)
	if err != nil {
		errLog.Fatalln(err)
	}
	defer s.Close()

	serverShutdownCh := make(chan struct{})
	server := &http.Server{
		ErrorLog: errLog,
		Addr:     config.SERVER_ADDR(),
	}
	server.RegisterOnShutdown(func() {
		close(serverShutdownCh)
	})
	server.Handler = handler.Compile(&handler.App{
		InfoLog:          infoLog,
		ErrLog:           errLog,
		ServerShutdownCh: serverShutdownCh,
		Storage:          s,
	})

	serverErrCh := make(chan error)
	go func() {
		defer close(serverErrCh)
		infoLog.Printf("Server listen on \"%s\"\n", server.Addr)
		serverErrCh <- server.ListenAndServe()
	}()

	signalChan := make(chan os.Signal, 1)
	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	signal.Notify(signalChan, signals...)

	select {
	case err := <-serverErrCh:
		errLog.Panicln("Server returning error: ", err)
	case sig := <-signalChan:
		signal.Reset(signals...)
		dur := 1*time.Minute + 30*time.Second
		infoLog.Printf(
			"Got \"%s\" signal, Stopping (Waiting for graceful shutdown: %s)\n",
			sig.String(), dur.String(),
		)
		ctx, cancel := context.WithTimeout(context.Background(), dur)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			errLog.Panicln("Shutting down server returning error: ", err)
		}
	}
}
