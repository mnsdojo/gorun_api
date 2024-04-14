package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type CompileRequest struct {
	Code string `json:"code"`
}

type CompileResult struct {
	Result string `json:"result"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	r.Post("/compile", compileHandler)

	http.ListenAndServe(":3000", r)
}

func compileHandler(w http.ResponseWriter, r *http.Request) {
	var req CompileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to parse JSON request body", http.StatusBadRequest)
		return
	}

	// Write code to a temporary file
	tempFile, err := os.CreateTemp("", "source*.go")
	if err != nil {
		http.Error(w, "Failed to create temporary file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := tempFile.WriteString(req.Code); err != nil {
		http.Error(w, "Failed to write code to temporary file", http.StatusInternalServerError)
		return
	}

	output, err := compileCode(tempFile.Name())
	if err != nil {
		http.Error(w, fmt.Sprintf("Compilation failed: %s", err), http.StatusInternalServerError)
		return
	}

	// Return compilation output as JSON response
	res := CompileResult{Result: output}
	json.NewEncoder(w).Encode(res)
}

func compileCode(filePath string) (string, error) {
	dir := filepath.Dir(filePath)

	cmd := exec.Command("go", "run", filePath)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
