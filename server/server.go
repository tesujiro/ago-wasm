package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

type server struct {
	router *http.ServeMux
}

func newServer() *server {
	return &server{
		router: http.NewServeMux(),
	}
}

func main() {
	p := flag.Int("p", 80, "port number")
	flag.Parse()
	addr := fmt.Sprintf(":%d", *p)

	s := newServer()
	s.routes()
	fmt.Println("Start listening on", addr)
	err := http.ListenAndServe(addr, s.router)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (s *server) routes() {
	s.router.HandleFunc("/", s.handleDefault())
	s.router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
}

func (s *server) handleDefault() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		fmt.Println("Default handler got a request.")
		http.Redirect(w, r, "/static", 301)
	}
}
