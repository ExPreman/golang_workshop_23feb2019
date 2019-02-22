package cors

import (
	"net/http"

	middleware "github.com/payfazz/go-middleware"
)

func OriginAll() middleware.Func {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next(w, r)
		}
	}
}
