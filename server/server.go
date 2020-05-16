package server

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/jatinsandilya/tez/redis"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Server holds all shared dependencies
// this service use
type Server struct {
	rp     *redis.Pool
	router *mux.Router
}

//Create creates a new server at startup
// of this service
func Create() *Server {

	s := &Server{
		rp:     redis.NewPool(),
		router: mux.NewRouter().StrictSlash(true),
	}

	s.router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	s.setupRoutes(s.router.PathPrefix("/v1").Subrouter())
	return s
}

//Shutdown stops the server closing all the connections it holds
// Along with any sanity action required
func (s *Server) Shutdown() {
	log.Println("Terminating.")
	s.rp.Close()
}

//Logger is a logging middleware function
// to wrap all handlers
func Logger(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		log.Println("Enterred handler : " + getFunctionName(h))
		defer log.Printf("Exiting handler : "+getFunctionName(h)+" . Time taken: %v", time.Now().Sub(startTime))
		w.Header().Set("Content-Type", "application/json")
		h(w, r)
	})
}

func getFunctionName(i interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	name = strings.Split(name, ".")[3]
	name = strings.Split(name, "-")[0]
	return name
}

//Serve starts to listen at specific port
func (s *Server) Serve() {
	srv := &http.Server{
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 100 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      s.router,
		Addr:         ":3000",
	}
	log.Println("Listening on port : 3000")
	log.Fatal(srv.ListenAndServe())
}
