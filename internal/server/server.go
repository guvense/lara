package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	lara "github.com/guvense/lara/internal"
	callback "github.com/guvense/lara/internal/callback"
	model "github.com/guvense/lara/internal/model"
	parser "github.com/guvense/lara/internal/parser"
)

type ConfigCORS struct {
	Methods          []string `yaml:"methods"`
	Headers          []string `yaml:"headers"`
	Origins          []string `yaml:"origins"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

type Server struct {
	MocPath    string
	Router     *mux.Router
	HttpServer *http.Server
	Config     lara.Config
}

func (s *Server) Build() error {
	if _, err := os.Stat(s.MocPath); os.IsNotExist(err) {
		log.Printf("%v: the directory %s doesn't exists", err, s.MocPath)
	}

	var filePathsCh = make(chan string)
	var done = make(chan bool)

	go func() {
		FindMocks(s.MocPath, filePathsCh)
		done <- true
	}()
loop:
	for {
		select {
		case filePath := <-filePathsCh:
			if filepath.Ext(filePath) != ".json" {
				continue
			}
			var mocs []model.Moc
			err := UnmarshalMocs(filePath, &mocs)
			if err != nil {
				log.Printf("error trying to load %s mocs: %v", filePath, err)
			} else {
				s.addMock(mocs)
			}
		case <-done:
			close(filePathsCh)
			close(done)
			break loop
		}
	}
	return nil
}

func (s *Server) addMock(mocs []model.Moc) {

	callback := callback.CallBack{
		TokenServers: s.Config.TokenGenerator.TokenServers,
	}

	for _, moc := range mocs {

		r := s.Router.HandleFunc(moc.Rest.Request.Endpoint, MocHandler(moc, callback, s)).
			Methods(moc.Rest.Request.Method)

		if moc.Rest.Request.Headers != nil {
			for k, v := range *moc.Rest.Request.Headers {
				r.HeadersRegexp(k, v)
			}
		}

		if moc.Rest.Request.Params != nil {
			for k, v := range *moc.Rest.Request.Params {
				r.Queries(k, v)
			}
		}
	}
}

func MocHandler(moc model.Moc, callback callback.CallBack, s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		requestBody, _ := io.ReadAll(r.Body)
		r.ParseForm()
		values := r.URL.Query()

		pathVariables := mux.Vars(r)
	
		queryParams := make(map[string]string)

		for key, val := range values {
			queryParams[key] = val[0]
		}

		pReq := parser.RequestParser {
			Body:json.RawMessage(requestBody),
			Params: &queryParams,
			PathVariables: &pathVariables,
			
		}

		pars := parser.Parser {
			Request: pReq,
			Config: s.Config,
		}
		
		stringJson := fmt.Sprintf("%v", string(moc.Rest.Response.Body))
		body := pars.Parse(stringJson)

		pRes := parser.ResponseParser{
			Body: json.RawMessage(body),
		}

		(&pars).Response = pRes

		if moc.Rest.Response.Delay.Delay() > 5 {
			time.Sleep(moc.Rest.Response.Delay.Delay())
		}
		writeHeaders(moc, w)
		w.WriteHeader(moc.Rest.Response.Status)
		writeBody(body, w)

		if len(moc.Rest.Callback) > 0 {
			go callback.Call(moc.Rest.Callback, &pars)
		}
	}
}

func writeHeaders(moc model.Moc, w http.ResponseWriter) {
	if moc.Rest.Response.Headers == nil {
		return
	}

	for key, val := range *moc.Rest.Response.Headers {
		w.Header().Set(key, val)
	}
}

func writeBody(body string, w http.ResponseWriter) {

	wb := []byte(body)
	w.Write(wb)
}

func (s *Server) runServer() error {

	return s.HttpServer.ListenAndServe()

}

func (s *Server) Run() {
	go func() {
		log.Printf("Lara is on tap now: %s\n", s.HttpServer.Addr)
		err := s.runServer()
		if err != http.ErrServerClosed {
			fmt.Print("Error occurred while initializing server: %w", err)
			log.Fatal(err)
		}
		fmt.Print("Server initilized")
	}()
}

func (s *Server) Shutdown() error {
	log.Println("stopping server...")
	if err := s.HttpServer.Shutdown(context.TODO()); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	return nil
}

var (
	defaultCORSMethods        = []string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE", "PATCH", "TRACE", "CONNECT"}
	defaultCORSHeaders        = []string{"X-Requested-With", "Content-Type", "Authorization"}
	defaultCORSExposedHeaders = []string{"Cache-Control", "Content-Language", "Content-Type", "Expires", "Last-Modified", "Pragma"}
)

func PrepareAccessControl(configCors ConfigCORS) (h []handlers.CORSOption) {
	h = append(h, handlers.AllowedMethods(defaultCORSMethods))
	h = append(h, handlers.AllowedHeaders(defaultCORSHeaders))
	h = append(h, handlers.ExposedHeaders(defaultCORSExposedHeaders))
	return
}
