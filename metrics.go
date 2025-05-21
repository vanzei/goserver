package main

import (
	"net/http"
	"html/template"
	"path/filepath"
)

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
    // Define the template path
    tmplPath := filepath.Join("templates", "metrics.html")
    
    // Parse the template
    tmpl, err := template.ParseFiles(tmplPath)
    if err != nil {
        http.Error(w, "Failed to load template", http.StatusInternalServerError)
        return
    }
    
    // Define data to pass to the template
    data := struct {
        Hits int32
    }{
        Hits: cfg.fileserverHits.Load(),
    }
    
    // Set content type
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    
    // Execute the template
    if err := tmpl.Execute(w, data); err != nil {
        http.Error(w, "Failed to render template", http.StatusInternalServerError)
        return
    }
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}