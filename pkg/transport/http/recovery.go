package http

import (
	"net/http"

	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
)

func Recovery(logger log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if rvr == http.ErrAbortHandler {
						// we don't recover http.ErrAbortHandler so the response
						// to the client is aborted, this should not be logged
						panic(rvr)
					}

					event := logger.Error().
						Str("method", request.Method).
						Str("url", request.RequestURI).
						Int("code", http.StatusInternalServerError)

					err, ok := rvr.(error)
					if ok {
						event = event.Err(err)
					}

					event.Msg("recovered from panic")

					if request.Header.Get("Connection") != "Upgrade" {
						writer.WriteHeader(http.StatusInternalServerError)
					}
				}
			}()

			next.ServeHTTP(writer, request)
		}

		return http.HandlerFunc(fn)
	}
}
