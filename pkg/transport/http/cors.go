package http

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Relay-Secret")

		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
