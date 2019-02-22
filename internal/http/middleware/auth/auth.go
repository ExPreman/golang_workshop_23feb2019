package auth

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	middleware "github.com/payfazz/go-middleware"
	"github.com/payfazz/go-middleware/common/kv"

	"github.com/win-t/golang_workshop_23feb2019/internal/storage"
)

type key string

const nameKey = key("name")
const isAdminKey = key("is_admin")

func New(errLog *log.Logger, s *storage.Storage) middleware.Func {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			func() {
				authPart := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
				if len(authPart) < 2 {
					return
				}

				if strings.ToLower(authPart[0]) != "basic" {
					return
				}

				rawPart, err := base64.StdEncoding.DecodeString(authPart[1])
				if err != nil {
					return
				}

				basicPart := strings.SplitN(string(rawPart), ":", 2)
				if len(basicPart) < 2 {
					return
				}
				user := basicPart[0]
				pass := basicPart[1]

				isAdmin, err := s.IsAdmin(r.Context(), user, pass)
				if err == storage.ErrNotFound {
					return
				}
				if err != nil {
					errLog.Println(err)
					return
				}

				kv.Set(r, nameKey, user)
				kv.Set(r, isAdminKey, isAdmin)

			}()
			next(w, r)
		}
	}
}

func Name(r *http.Request) string {
	x := kv.Get(r, nameKey)
	if x == nil {
		return ""
	}
	return x.(string)
}

func IsAdmin(r *http.Request) bool {
	x := kv.Get(r, isAdminKey)
	if x == nil {
		return false
	}
	return x.(bool)
}
