package pkg

import "net/http"

type message struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewMessage(w http.ResponseWriter, code int, msg string) error {
	return WriteJson(w, code, message{
		Code: code,
		Msg:  msg,
	})
}
