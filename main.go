package main

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/xperimental/bukky/internal/memory"
	"github.com/xperimental/bukky/internal/web"
)

const (
	envAddr = "LISTEN_ADDR"
)

var (
	addr = ":8080"

	log = &logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			DisableTimestamp: true,
		},
		Hooks:    logrus.LevelHooks{},
		Level:    logrus.InfoLevel,
		ExitFunc: os.Exit,
	}
)

func main() {
	if value, ok := os.LookupEnv(envAddr); ok {
		addr = value
	}

	store := memory.NewStore(log)
	r := web.NewRouter(log, store)

	log.Infof("Listening on %s ...", addr)
	if err := http.ListenAndServe(addr, r.Handler()); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
