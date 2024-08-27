package pprofile

import (
	"expvar"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/metrics"
)

type ppServerT struct {
	path     string
	http     string
	abstract string
	mux      *http.ServeMux
	unixSrv  *http.Server
	httpSrv  *http.Server
	absSrv   *http.Server
}

var Server *ppServerT

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

	s := &ppServerT{
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
	if disable == "1" || disable == "true" || disable == "TRUE" || EnableOnInit != "1" {
		return
	}

	Server.ServeHTTP()
	Server.ServeSocket()
	Server.ServeAbstract()
}

func (s *ppServerT) ServeHTTP() {

	if dbglog() {
		log.Printf("DBG: http=[%s]", s.path)
	}

	if s.http == "" {
		return
	}

	if dbglog() {
		log.Printf("DBG: %s", s.http)
	}

	l, err := net.Listen(HTTP_NETWORK, s.http)
	if err != nil {
		// XXX - logging
		return
	}
	s.httpSrv = &http.Server{
		Handler: s.mux,
	}
	go s.httpSrv.Serve(l)

}

func (s *ppServerT) ServeSocket() {

	if dbglog() {
		log.Printf("DBG: unix-socket=[%s]", s.path)
	}
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
	s.unixSrv = &http.Server{
		Handler: s.mux,
	}
	go s.unixSrv.Serve(l)
}

func (s *ppServerT) ServeAbstract() {

	if dbglog() {
		log.Printf("DBG: unix-abstract=[%s]", s.abstract)
	}

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

	s.absSrv = &http.Server{
		Handler: s.mux,
	}

	go s.absSrv.Serve(l)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	descs := metrics.All()
	samples := make([]metrics.Sample, len(descs))
	for i := range samples {
		samples[i].Name = descs[i].Name
	}
	metrics.Read(samples)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	raw := r.URL.Query().Has("raw")

	fmt.Fprintf(w, "{\n")
	first := ""
	for _, sample := range samples {
		name, value := sample.Name, sample.Value
		// Handle each sample.
		switch value.Kind() {
		case metrics.KindUint64:
			fmt.Fprintf(w, `%s"%s": %d`, first, name, value.Uint64())
			first = ",\n"
		case metrics.KindFloat64:
			fmt.Fprintf(w, `%s"%s": %f`, first, name, value.Float64())
			first = ",\n"
		case metrics.KindFloat64Histogram:
			if !raw {
				continue
			}
			fmt.Fprintf(w, `%s"%s": `, first, name)
			h := value.Float64Histogram()
			fmt.Fprintf(w, `{`)
			fst := ""
			for i := range value.Float64Histogram().Counts {
				fmt.Fprintf(w, `%s"%f": %d`, fst, h.Buckets[i], h.Counts[i])
				fst = ","
			}
			fmt.Fprintf(w, `}`)
		case metrics.KindBad:
		default:
		}
	}
	fmt.Fprintf(w, "\n}\n")
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

func Handler(r *http.Request) (h http.Handler, pattern string) {
	return Server.mux.Handler(r)
}

func Handle(pattern string, handler http.Handler) {
	Server.mux.Handle(pattern, handler)
}

func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	Server.mux.HandleFunc(pattern, handler)
}
