package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

func main() {
	mux := http.NewServeMux()
	apcfg := &apiConfig{}
	fileserver := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("GET /app/", apcfg.middlewareMetricsInc(fileserver))
	mux.HandleFunc("GET /admin/metrics/", apcfg.metrics)
	mux.HandleFunc("POST /admin/reset", apcfg.metricsResetHandler)
	mux.HandleFunc("GET /api/healthz", healthHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return
	}

}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (a *apiConfig) metricsResetHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	a.fileserverHits.Store(0)
	_, err := w.Write([]byte("Reset"))
	if err != nil {
		return
	}
}

func (a *apiConfig) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fileContent, err := os.ReadFile("metrics.html")
	_, err = w.Write([]byte(fmt.Sprintf(string(fileContent), a.fileserverHits.Load())))
	if err != nil {
		return
	}
}
