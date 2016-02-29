package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var cfg *Config

func main() {
	err := readConfig()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.CertFile == "" {
		log.Println("Starting HTTP server on ", cfg.ListenAddr)
		err = http.ListenAndServe(cfg.ListenAddr, GalleryHandler{})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Starting HTTPS server on ", cfg.ListenAddr)
		err = http.ListenAndServeTLS(cfg.ListenAddr, cfg.CertFile, cfg.KeyFile, GalleryHandler{})
		if err != nil {
			log.Fatal(err)
		}
	}

}

type GalleryHandler struct{}

func (h GalleryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if cfg.BasicAuthUser != "" && !checkBasicAuth(w, r) {
		return
	}

	switch r.URL.Path {
	case "", "/":
		index(w, r)
	case "/file":
		serveFile(w, r)
	case "/image":
		serveImage(w, r)
	default:
		serveGallery(w, r)
	}
}

type Config struct {
	ListenAddr    string
	BaseURL       string
	CertFile      string
	KeyFile       string
	BasicAuthUser string
	BasicAuthPass string
}

func readConfig() error {
	b, err := ioutil.ReadFile(filepath.Join(filepath.Dir(os.Args[0]), configFile))
	if err != nil {
		return err
	}
	cfg = new(Config)
	err = json.Unmarshal(b, cfg)
	if err != nil {
		return err
	}
	return nil
}

func checkBasicAuth(w http.ResponseWriter, r *http.Request) bool {
	u, p, ok := r.BasicAuth()
	if ok && u == cfg.BasicAuthUser && p == cfg.BasicAuthPass {
		return true
	}
	w.Header().Add("Www-Authenticate", "Basic")
	w.WriteHeader(http.StatusUnauthorized)
	return false
}
