package server

import (
	"go_final_project/pkg/api"
	"log"
	"net/http"
	"os"
)

const (
	defaultPort = "7540"
	webDir      = "./web"
)

func Run() error {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = defaultPort
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	api.Init(mux)

	log.Printf("Сервер запущен на порту %s", port)
	return http.ListenAndServe(":"+port, mux)
}
