package main

import (
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/apps_scootsy/raas/server"
	log "github.com/sirupsen/logrus"
)

// @title Redis as a Service - API
// @version 1.0
// @description This is the service to store/retrieve data in Redis. To be used as a cache.

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /
func main() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	s := server.Create()

	go func() {
		sig := <-sigs
		log.Println("Recieved signal - " + sig.String())
		s.Shutdown()
		os.Exit(1)
	}()

	s.Serve()
}
