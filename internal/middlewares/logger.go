package middlewares

import (
	"log"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler in the chain
		next.ServeHTTP(w, r)

		log.Printf(
			"%s %s %s\n",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}
