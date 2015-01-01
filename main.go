package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/NickPresta/chameleon/cache"
	"github.com/NickPresta/chameleon/config"
	"github.com/NickPresta/chameleon/handlers"
)

var configPath *string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get current working directory")
	}
	configFile := path.Join(cwd, "config.json")
	configPath = flag.String("config", configFile, "Full path to configuration file")
}

func main() {
	flag.Parse()
	servers, err := config.ParseConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	for _, server := range servers {
		log.Printf("Starting proxy for '%v'\n", server.URL)
		wg.Add(1)
		go func(s config.ServerDefinition) {
			defer wg.Done()

			cacher := cache.NewDiskCacher(s.DataDirectory)

			mux := http.NewServeMux()

			mux.Handle("/", handlers.CachedProxyMiddleware(handlers.ProxyHandler, s, cacher))
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), mux))
		}(server)
	}
	wg.Wait()
}
