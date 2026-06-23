package main

import (
	"io"
	"net/http"
)

type App struct {
	store *Store
}

func NewServer(store *Store) *http.Server {
	server := &http.Server{Addr: ":8000"}

	app := &App{store}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome!"))
	}))
	mux.Handle("GET /{key}", http.HandlerFunc(app.Get))
	mux.Handle("POST /{key}", http.HandlerFunc(app.Set))
	mux.Handle("DELETE /{key}", http.HandlerFunc(app.Delete))

	server.Handler = mux

	return server
}

func (a *App) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	println("getting: key: " + key)
	if data, ok := a.store.Get(key); !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {

		w.Write([]byte(data))
	}

}

func (a *App) Set(w http.ResponseWriter, r *http.Request) {
	if data, err := io.ReadAll(r.Body); err != nil {
		println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		a.store.Set(r.PathValue("key"), data)
		w.Write([]byte("ok"))
	}

}

func (a *App) Delete(w http.ResponseWriter, r *http.Request) {
	a.store.Delete(r.PathValue("key"))

	w.Write([]byte("deleted!"))

}
