// HTTP interface for the store
package main

import (
	"io"
	"net/http"
)

type App struct {
	s *Store
}

func NewServer(s *Store) *http.Server {
	app := &App{s}
	server := &http.Server{Addr: "8000"}
	mux := http.NewServeMux()
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome !"))
	}))

	mux.Handle("GET /{key}", http.HandlerFunc(app.get))
	mux.Handle("POST /{key}", http.HandlerFunc(app.set))
	mux.Handle("DELETE /{key}", http.HandlerFunc(app.delete))

	server.Handler = mux

	return server
}

func (a *App) get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if data, ok := a.s.Get(key); !ok {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Write(data)
	}

}

func (a *App) set(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if data, err := io.ReadAll(r.Body); err != nil {
		println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		a.s.Set(key, data)
		w.Write([]byte("ok"))
	}
}

func (a *App) delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	a.s.Delete(key)

	w.Write([]byte("key: " + key + " Deleted!"))

}
