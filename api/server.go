package router

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	Router *http.ServeMux
}

func New() *Server {
	router := http.NewServeMux()
	s := &Server{
		Router: router,
	}
	return s
}

func (s *Server) Run(port string) error {
	server := http.Server{
		Addr:    port,
		Handler: s.Router,
	}

	fmt.Printf("Listening on %s", port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
		return err
	}

	return nil
}
