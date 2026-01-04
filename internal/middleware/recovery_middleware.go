package middleware

import (
	"log"
	"net/http"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC RECOVERED: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Gatekeeper: An internal error occurred"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
