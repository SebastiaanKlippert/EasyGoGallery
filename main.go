package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

var cfg *Config

func main() {
	err := readConfig()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_RDWR|os.O_APPEND, os.ModeAppend)
		if err != nil {
			f, err = os.Create(cfg.LogFile)
		}
		if err != nil {
			log.Println(err)
		}
		if f != nil {
			log.Printf("Output set to %s", cfg.LogFile)
			log.SetOutput(f)
		}
	}

	log.Printf("Using %d of %d CPUs", runtime.GOMAXPROCS(0), runtime.NumCPU())

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
	LogFile       string
	CertFile      string
	KeyFile       string
	BasicAuthUser string
	BasicAuthPass string
	LogAuth       bool
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
		//logAuth(r, true, &u, &p)
		return true
	}
	w.Header().Add("Www-Authenticate", "Basic")
	w.WriteHeader(http.StatusUnauthorized)
	logAuth(r, false, &u, &p)
	return false
}

func logAuth(r *http.Request, allowed bool, u, p *string) {
	if !cfg.LogAuth {
		return
	}
	ok, creds := "DENIED", ""
	if allowed {
		ok = "ALLOWED"
	} else {
		creds = fmt.Sprintf(" Used credentials (user - pass): %s - %s", *u, *p)
	}

	log.Printf("[AUTH] Access %s for %s on %s%s\n", ok, r.RemoteAddr, r.URL.Path, creds)
}
