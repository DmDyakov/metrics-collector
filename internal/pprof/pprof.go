package pprof

import (
	"net/http"
	_ "net/http/pprof"
)

// Serve запускает pprof-сервер на указанном адресе.
func Serve(addr string) error {
	return http.ListenAndServe(addr, nil)
}
