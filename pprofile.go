package pprofile

import (
	"expvar"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

type ppServerT struct {
	path     string
	http     string
	abstract string
	mux      *http.ServeMux
}

var Server *ppServerT

func Handler(r *http.Request) (h http.Handler, pattern string) {
	return Server.mux.Handler(r)
}

func Handle(pattern string, handler http.Handler) {
	Server.mux.Handle(pattern, handler)
}

func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	Server.mux.HandleFunc(pattern, handler)
}

func init() {
	if v, ok := os.LookupEnv("PPROF_ENABLE_HTTP"); ok {
		EnableHTTP = v
	}
	if v, ok := os.LookupEnv("PPROF_ENABLE_SOCKET"); ok {
		EnableTmpSocket = v
	}
	if v, ok := os.LookupEnv("PPROF_ENABLE_ABSTRACT"); ok {
		EnableAbstractSocket = v
	}

	if v, ok := os.LookupEnv("PPROF_HTTP_LISTEN"); ok {
		ListenHTTP = v
	}

	if v, ok := os.LookupEnv("PPROF_SOCKET_LISTEN"); ok {
		TmpSocketTemplate = v
	}
	if v, ok := os.LookupEnv("PPROF_ABSTRACT_LISTEN"); ok {
		AbstractSocketTemplate = v
	}

	s := ppServerT{
		mux: Mux(),
	}

	if EnableHTTP == "1" || EnableHTTP == "on" || EnableHTTP == "true" {
		s.http = ListenHTTP
	}

	if EnableTmpSocket == "1" || EnableTmpSocket == "on" || EnableTmpSocket == "true" {
		s.path = Expand(TmpSocketTemplate)
	}

	if EnableAbstractSocket == "1" || EnableAbstractSocket == "on" || EnableAbstractSocket == "true" {
		s.abstract = Expand(AbstractSocketTemplate)
	}

	FailOnError = os.Getenv("PPROF_FAIL")
	DebugLog = os.Getenv("PPROF_DEBUG")

	Server = s
	disable := os.Getenv("PPROF_OFF")
	if disable == "1" || disable == "true" || disable == "TRUE" {
		return
	}

	Server.ServeHTTP()
	Server.ServeSocket()
	Server.ServeAbstract()
}

func (s *ppServerT) ServeHTTP() {

	if s.http == "" {
		return
	}

	if dbglog() {
		log.Printf("DBG: %s", s.http)
	}
	go http.ListenAndServe(s.http, s.mux)
}

func (s *ppServerT) ServeSocket() {

	if s.path == "" {
		return
	}

	// Ensure existing socket is not listening.
	conn, err := net.Dial("unix", s.path)
	if err != nil {
		os.Remove(s.path)
	} else {
		conn.Close()
		return
	}

	if dbglog() {
		log.Printf("DBG: %s", s.path)
	}

	// Serve the handlers.
	l, err := net.Listen("unix", s.path)
	if err != nil {
		if dbglog() {
			log.Printf("[ERROR] %v", err)
		}
		if fail() {
			panic(err)
		}
		return
	}
	os.Chmod(s.path, 0666)
	server := &http.Server{
		Handler: s.mux,
	}
	go server.Serve(l)
}

func (s *ppServerT) ServeAbstract() {

	if s.abstract == "" {
		return
	}

	if dbglog() {
		log.Printf("DBG: %s", s.abstract)
	}

	l, err := net.Listen("unix", s.abstract)
	if err != nil {
		if dbglog() {
			log.Printf("[ERROR] %v", err)
		}
		if fail() {
			panic(err)
		}
		return
	}

	server := &http.Server{
		Handler: s.mux,
	}

	go server.Serve(l)
}

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}

func Cleanup() {
	if Server.path != "" {
		os.Remove(Server.path)
	}
}
