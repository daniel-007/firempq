package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/vburenin/firempq/conf"
	"github.com/vburenin/firempq/fctx"
	"github.com/vburenin/firempq/log"
	"github.com/vburenin/firempq/server"
)

func main() {
	// Initialize logging to a default INFO level to be able to log config error.
	log.InitLogging()
	conf.ParseConfigParameters()
	if len(conf.CFG.Profiler) > 0 {
		go func() {
			if err := http.ListenAndServe(conf.CFG.Profiler, nil); err != nil {
				log.Error("Could not initialize profiler: %v", err)
			}
		}()
	}

	// Reinitialize log level according to the config data.
	log.InitLogging()
	ctx := fctx.Background("start")
	server := server.NewServer(ctx)
	server.Start()
}
