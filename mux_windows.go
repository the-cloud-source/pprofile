//go:build windows
// +build windows

package pprofile

import (
	"net/http"
	"net/http/pprof"
)

func Mux() *http.ServeMux {

	mux := http.NewServeMux()

	// Export pprof.
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))

	// Export debugging vars.
	mux.Handle("/debug/vars", http.HandlerFunc(expvarHandler))

	return mux
}
