package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/win-t/golang_workshop_23feb2019/internal/config"
	"github.com/win-t/golang_workshop_23feb2019/internal/logger"
	"github.com/win-t/golang_workshop_23feb2019/internal/storage"
)

func main() {
	env := flag.Bool("env", false, "Print config from env and exit")
	flag.Parse()

	if *env {
		config.PrintAll()
		return
	}

	infoLog, errLog := logger.Open(!config.DISABLE_PLAIN_LOG())

	if len(flag.Args()) != 0 {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	closed := false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		signalChan := make(chan os.Signal, 1)
		signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT}
		signal.Notify(signalChan, signals...)
		sig := <-signalChan
		signal.Reset(signals...)
		infoLog.Printf(
			"Got \"%s\" signal, Stopping\n",
			sig.String(),
		)
		closed = true
		cancel()
	}()

	closeCh := ctx.Done()

	s, err := storage.Open(config.PG_URL(), errLog)
	if err != nil {
		errLog.Fatalln(err)
	}
	defer s.Close()

	w := s.NewWorkerConnection()
	defer w.Close()

	notifCh := w.NotifCh()

	infoLog.Println("worker is running")

LOOP:
	for !closed {
		workID, data, err := w.TakeJob(ctx)
		if err != nil {
			if err == storage.ErrNotFound {
				select {
				case <-notifCh:
					continue LOOP
				case <-closeCh:
					break LOOP
				}
			} else {
				errLog.Println(err)
				sleepOrClosed(5*time.Second, closeCh)
				continue LOOP
			}
		}

		infoLog.Printf("Got work %d with data %d\n", workID, data)

		total, err := s.GetTotalCounter(ctx)
		if err != nil {
			errLog.Println(err)
			if err := w.UntakeJob(ctx, workID); err != nil {
				errLog.Println(err)
				continue LOOP
			}
		}

		var score int64 = -1
		if total != 0 {
			score = (data * 100) / total
		}
		sleepOrClosed(5*time.Second, closeCh) // simulate hardwork

		if closed {
			func() {
				infoLog.Printf("work %d is incomplete due to worker is exiting\n", workID)
				untakeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := w.UntakeJob(untakeCtx, workID); err != nil {
					errLog.Println(err)
				} else {
					infoLog.Println("Debug untake")
				}
			}()
			break LOOP
		}

		if err := w.PublishJobResult(ctx, workID, score); err != nil {
			errLog.Println(err)
		}

		infoLog.Printf("work %d done with score %d\n", workID, score)
	}

	infoLog.Println("Exiting worker")
}

func sleepOrClosed(dur time.Duration, closeCh <-chan struct{}) {
	timer := time.NewTimer(dur)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-closeCh:
	}
}
