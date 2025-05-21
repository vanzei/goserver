package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	// Reset the database or perform any necessary cleanup
	// This is a placeholder implementation

	if cfg.PLATFORM != "dev" {
		http.Error(w, "Reset is only allowed in dev environment", http.StatusForbidden)
		return
	}
	if err := cfg.DB.ResetDatabase(r.Context()); err != nil {
		http.Error(w, "Failed to reset database", http.StatusInternalServerError)
		return
	}

	cfg.fileserverHits.Store(0)
	cfg.DB.ResetDatabase(r.Context())

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Database reset successfully"))
}