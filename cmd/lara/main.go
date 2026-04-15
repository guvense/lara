package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	lara "github.com/guvense/lara/internal"
	editorapi "github.com/guvense/lara/internal/editorapi"
	mocwatcher "github.com/guvense/lara/internal/mocwatcher"
	server "github.com/guvense/lara/internal/server"

	_ "embed"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/radovskyb/watcher"
)

const (
	_defaultConfigFile = ""
	_defaultHost       = "localhost"
	_defaultPort       = 8898
	_defaultMocksPath  = "/mocs"
	_defaultWatcher    = false
	_defaultConfigPath = "./config.yml"
)

func main() {

	var (
		configFilePath = flag.String("config", _defaultConfigFile, "configuration file path")
		host           = flag.String("host", _defaultHost, "host")
		port           = flag.Int("port", _defaultPort, "port")
		mocksPath      = flag.String("mocks", _defaultMocksPath, "mocks folder path")
		watcher        = flag.Bool("watcher", _defaultWatcher, "refresh server when mocs changed")
	)
	flag.Parse()
	
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	watchConfigPath := effectiveConfigPath(*configFilePath)

	config, err := lara.PrepareConfig(*host, *port, *watcher, *mocksPath, *configFilePath)

	if err != nil {
		log.Println(err)
	}

	if err := ensureMocksDirectory(config.MocksPath); err != nil {
		log.Fatal(err)
	}
	if err := ensureConfigFile(watchConfigPath); err != nil {
		log.Fatal(err)
	}

	s := runServer(config, watchConfigPath)
	s.Run()

	w := runWatcher(&s, config, watchConfigPath)
	
	<-done
	close(done)

	if w != nil {
		w.Close()
	}

	if err := s.Shutdown(); err != nil {
		log.Fatal(err)
	}

}

func runWatcher(currentSrv *server.Server, config lara.Config, configPath string) *watcher.Watcher {
	if !config.Watcher {
		return nil
	}

	if err := ensureConfigFile(configPath); err != nil {
		log.Fatal(err)
	}
	if err := ensureMocksDirectory(config.MocksPath); err != nil {
		log.Fatal(err)
	}

	w, err := mocwatcher.InitializeWatcher(config.MocksPath,configPath)
	if err != nil {
		log.Fatal(err)
	}
	mocwatcher.AttachWatcher(w, func() {
		if err := currentSrv.Shutdown(); err != nil {
			log.Fatal(err)
		}
		*currentSrv = runServer(config, configPath)
		currentSrv.Run()
	})
	return w
}

func runServer(config lara.Config, configPath string) server.Server {

	router := mux.NewRouter()
	httpAddr := fmt.Sprintf("%s:%d", config.ServerConfig.Host, config.ServerConfig.Port)

	cors := server.ConfigCORS{}
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: handlers.CORS(server.PrepareAccessControl(cors)...)(router),
	}

	server := server.Server{
		MocPath:    config.MocksPath,
		Router:     router,
		HttpServer: httpServer,
		Config:     config,
	}

	editorapi.AttachRoutes(router, editorapi.Options{
		ConfigPath: configPath,
		MocksPath:  config.MocksPath,
	})

	if err := server.Build(); err != nil {
		log.Fatal(err)
	}
	return server
}

func effectiveConfigPath(path string) string {
	if path != "" {
		return path
	}
	return _defaultConfigPath
}

func ensureConfigFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte("\n"), 0o600)
}

func ensureMocksDirectory(path string) error {
	return os.MkdirAll(path, 0o755)
}
