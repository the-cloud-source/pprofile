//go:build darwin || dragonfly || freebsd || linux || nacl || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

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

	// Export /proc.
	fs := http.FileServer(http.Dir("/proc/self"))
	mux.Handle("/debug/proc/", http.StripPrefix("/debug/proc/", fs))

	// Export debugging vars.
	mux.Handle("/debug/vars", http.HandlerFunc(expvarHandler))

	return mux
}
