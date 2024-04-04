package main

import (
	"context"
	"log"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(port string, handler http.Handler) error {
	s.httpServer = &http.Server{Addr: ":" + port, Handler: handler}
	log.Printf("Запуск сервера на порте: %s", port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Остановка сервера")
	return s.httpServer.Shutdown(ctx)
}