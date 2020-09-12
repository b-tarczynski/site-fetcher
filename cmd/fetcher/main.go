package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BarTar213/bartlomiej-tarczynski/api"
	"github.com/BarTar213/bartlomiej-tarczynski/config"
	"github.com/BarTar213/bartlomiej-tarczynski/storage"
)

func main() {
	conf := config.NewConfig("fetcher.yml")
	logger := log.New(os.Stdout, "", log.LstdFlags)

	logger.Printf("%+v\n", conf)

	postgres, err := storage.NewPostgres(&conf.Postgres)
	if err != nil {
		logger.Fatalf("Postgres error: %s", err)
	}

	a := api.NewApi(
		api.WithConfig(conf),
		api.WithLogger(logger),
		api.WithStorage(postgres),
	)

	go a.Run()
	logger.Print("started app")

	shutDownSignal := make(chan os.Signal)
	signal.Notify(shutDownSignal, syscall.SIGINT, syscall.SIGTERM)

	<-shutDownSignal
	logger.Print("exited from app")
}