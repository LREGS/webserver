package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	r := chi.NewRouter()

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)

	api := chi.NewRouter()
	r.Mount("/api", api)

	api.Get("/healthz", handlerReadiness)
	api.Get("/reset", apiCfg.handlerReset)
	api.Post("/validate_chirp", chirpValidator)

	adminApi := chi.NewRouter()
	r.Mount("/admin", adminApi)

	adminApi.Get("/metrics", apiCfg.handlerMetrics)

	corsMux := middlewareCors(r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

const htmlMessage = `<html>

<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>`

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(htmlMessage, cfg.fileserverHits)))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func chirpValidator(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVal struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Couldn't load parameters")
		return
	}

	const maxLength = 140
	if len(params.Body) > maxLength {
		RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	if wordReplacer(params.Body) != "" {
		log.Printf(wordReplacer(params.Body))
		respondWithJSON(w, http.StatusOK, returnVal{
			Cleaned_body: wordReplacer(params.Body),
		})
	} else {
		log.Printf(params.Body)
		respondWithJSON(w, http.StatusOK, returnVal{
			Cleaned_body: params.Body,
		})
	}

}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Respondinf with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func wordReplacer(chirp string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Fields(chirp)
	words_found := 0

	for i, word := range words {
		for _, bWord := range badWords {
			// var replacement = ""
			if strings.ToLower(word) == bWord {
				// replacement += muteGenerator(len(word))
				words[i] = "****"
				words_found++
				break
			}

		}
	}

	if words_found > 0 {
		cleanMessage := strings.Join(words, " ")
		return cleanMessage
	} else {
		return ""
	}

}

func muteGenerator(count int) string {
	mute := ""
	for i := 0; i < count; i++ {
		mute += "*"
	}
	return mute
}
