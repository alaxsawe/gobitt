// Gobitt is an implementation of a file tracker, for the bittorent
// p2p protocol.
package gobitt

import (
	"github.com/jbonachera/gobitt/tracker"
	"github.com/jbonachera/gobitt/tracker/config"
	"github.com/jbonachera/gobitt/tracker/context"
	"github.com/jbonachera/gobitt/tracker/plugin"
	_ "github.com/jbonachera/gobitt/tracker/plugin/database"
	"log"
	"net/http"
	"time"
)

type contextFunc func(c context.ApplicationContext, w http.ResponseWriter, r *http.Request)

func (h contextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.f(h.c, w, r)
}

type contextHandler struct {
	c context.ApplicationContext
	f contextFunc
}

func purgePeersRunner(context context.ApplicationContext, maxAge time.Duration, loop time.Duration) {
	for {
		context.Database.PurgePeers(maxAge)
		time.Sleep(loop)
	}
}

func Start() {
	cfg := config.GetConfig()
	context := context.ApplicationContext{}
	context.Database = plugin.GetDatabasePlugin(cfg.Server.DatabasePlugin)
	context.Database.Start()

	log.Print("Starting background database cleaner")
	go purgePeersRunner(context, 10*time.Second, 30*time.Second)

	log.Print("Running on: " + cfg.Server.BindAddress + ":" + cfg.Server.Port)
	http.Handle("/announce", contextHandler{context, tracker.AnnounceHandler})
	http.Handle("/scrape", contextHandler{context, tracker.ScrapeHandler})
	http.ListenAndServe(cfg.Server.BindAddress+":"+cfg.Server.Port, nil)
}
