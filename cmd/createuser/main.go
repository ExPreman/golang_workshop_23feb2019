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
	env := flag.Bool("env", false, "Print config from env and exit")
	user := flag.String("user", "", "username")
	pass := flag.String("pass", "", "password")
	admin := flag.Bool("admin", false, "create user as admin user")
	flag.Parse()

	if *env {
		config.PrintAll()
		return
	}

	infoLog, errLog := logger.Open(!config.DISABLE_PLAIN_LOG())

	if len(flag.Args()) != 0 || *user == "" || *pass == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	s, err := storage.Open(config.PG_URL(), errLog)
	if err != nil {
		errLog.Fatalln(err)
	}
	defer s.Close()

	if *admin {
		if err := s.AddAdmin(context.Background(), *user, *pass); err != nil {
			errLog.Fatalln(err)
		}
	} else {
		if err := s.AddUser(context.Background(), *user, *pass); err != nil {
			errLog.Fatalln(err)
		}
	}

	infoLog.Println("Done")
}
