//go:build darwin || dragonfly || freebsd || linux || nacl || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package pprofile

import (
	"net/http"
	"net/http/pprof"

	"github.com/the-cloud-source/pprofile/fs"
)

func Mux() *http.ServeMux {

	mux := http.NewServeMux()

	// Export pprof.
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	// Exports /proc.
	f := fs.FileServer(fs.Dir("/proc/self"))
	mux.Handle("/debug/proc/", fs.StripPrefix("/debug/proc/", f))

	// Export debugging vars.
	mux.Handle("/debug/vars", http.HandlerFunc(expvarHandler))
	mux.Handle("/debug/metrics", http.HandlerFunc(metricsHandler))

	return mux
}
