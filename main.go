package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	fileserver := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("GET /app/", fileserver.ServeHTTP)
	mux.HandleFunc("GET /healthz", healthHandler)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return
	}
}
