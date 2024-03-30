package http

import (
	"net/http"
	"strings"

	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
)

const (
	HeaderKey = "Authorization"
)

func MiddlewareBasicAuthentication(username, password string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authUsername, authPassword, ok := r.BasicAuth()
		if ok && authUsername == username && authPassword == password {
			// delete the token from header, to avoid leak to the next server
			delHeader(HeaderKey, r)
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="Enter username and password"`)
		w.WriteHeader(http.StatusUnauthorized)
		handlerUtils.WriteResponse(w, []byte(`401 Unauthorized`))
	})
}

// Deletes the header, header name case insensitive
func delHeader(headerName string, r *http.Request) {
	headerName = strings.ToLower(headerName)
	for nme := range r.Header {
		if strings.ToLower(nme) == headerName {
			r.Header.Del(nme)
		}
	}
}
