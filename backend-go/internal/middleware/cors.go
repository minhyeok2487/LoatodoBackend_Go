package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

// NewCORS creates a CORS middleware handler.
func NewCORS() func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://loatodo.com", "https://www.loatodo.com", "http://localhost:3000", "http://localhost:3002"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	return c.Handler
}
