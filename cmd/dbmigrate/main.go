package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/win-t/golang_workshop_23feb2019/internal/config"
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

	if err := s.DoMigration(context.Background()); err != nil {
		errLog.Fatalln(err)
	}

	infoLog.Println("Done")
}
