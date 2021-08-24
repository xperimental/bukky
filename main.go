package main

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/xperimental/bukky/internal/memory"
	"github.com/xperimental/bukky/internal/web"
)

var (
	addr = ":8080"

	log = &logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			DisableTimestamp: true,
		},
		Hooks:        logrus.LevelHooks{},
		Level:        logrus.InfoLevel,
		ExitFunc:     os.Exit,
	}
)

func main() {
	store := memory.NewStore(log)
	r := web.MainHandler(log, store)

	log.Infof("Listening on %s ...", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
