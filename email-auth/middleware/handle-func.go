package middleware

import (
	"net/http"

	pkg "email-auth/pkg"
)

func HandleFunc(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			pkg.NewMessage(w, http.StatusBadRequest, err.Error())
			return
		}
	}
}
