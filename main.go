package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

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

type messageToValidate struct {
	Body string `json:"body"`
}

/*
type validatorResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error"`
}


func validateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var message messageToValidate
	err := decoder.Decode(&message)
	response := validatorResponse{}
	if err == nil && message.Body != "" {
		if len(message.Body) > 140 {
			w.WriteHeader(http.StatusBadRequest)
			response.Valid = false
			response.Error = "Chirp is too long"

		} else {
			w.WriteHeader(http.StatusOK)
			response.Valid = true
		}

	} else {
		response.Valid = false
		response.Error = "Something went wrong"
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	dat, _ := json.Marshal(response)
	_, err = w.Write(dat)
	if err != nil {
		return
	}
}
*/

func validateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var message messageToValidate
	err := decoder.Decode(&message)
	if err != nil || message.Body == "" || len(message.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
	}
	cleanMessage := replaceBadWords(message.Body)
	var response struct {
		CleanedBody string `json:"cleaned_body"`
	}
	response.CleanedBody = cleanMessage
	dat, err := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(dat)
	if err != nil {
		return
	}
}

func replaceBadWords(text string) string {
	newText := strings.Split(text, " ")
	for i, word := range newText {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			newText[i] = "****"
		}

	}
	return strings.Join(newText, " ")
}
