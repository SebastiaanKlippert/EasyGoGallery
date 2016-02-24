package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	//TODO config
	baseURL      = "http://thuis.klippert.nl:6543"
	galleryPath  = "./galleries"
	templatePath = "./templates"
	thumbSize    = 120
)

func serveGallery(w http.ResponseWriter, r *http.Request) {

	//Create gallery object
	g := new(Gallery)
	g.BaseURL = baseURL
	g.URLPath = strings.Trim(r.URL.Path, `/\`)
	g.ThumbSize = thumbSize
	g.ImagePath = filepath.Join(filepath.Dir(os.Args[0]), galleryPath, r.URL.Path)
	g.ThumbPath = filepath.Join(g.ImagePath, fmt.Sprintf("thumbs%d", g.ThumbSize))

	//List thumbs
	//TODO list images and create missing thumbs
	g.listImages(true)

	//Add template functions
	fm := template.FuncMap{
		"ThumbURL": g.ThumbURL,
	}

	//Parse template
	tmpl, err := template.New("gallery.html").Funcs(fm).ParseFiles(filepath.Join(filepath.Dir(os.Args[0]), templatePath, "gallery.html"))
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	//Render template
	err = tmpl.Execute(w, g)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	return
}

type Gallery struct {
	Error error

	BaseURL string
	URLPath string

	ImagePath string
	Images    []string

	ThumbSize int
	ThumbPath string
	Thumbs    []string

	collectThumbs bool
}

func (g *Gallery) listImages(thumbs bool) {
	g.collectThumbs = thumbs
	path := g.ImagePath
	if thumbs {
		path = g.ThumbPath
	}
	g.Error = filepath.Walk(path, g.collectImages)
}

func (g *Gallery) collectImages(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() && path != g.ImagePath && path != g.ThumbPath { //skip all subdirectories
		return filepath.SkipDir
	}
	name := info.Name()
	if !isPictureExt(filepath.Ext(name)) {
		return nil
	}
	if g.collectThumbs {
		g.Thumbs = append(g.Thumbs, name)
	} else {
		g.Images = append(g.Thumbs, name)
	}
	return nil
}

func (g *Gallery) ThumbURL(img string) string {
	uv := make(url.Values)
	uv.Add("gallery", g.URLPath)
	uv.Add("file", img)
	uv.Add("size", strconv.Itoa(g.ThumbSize))
	return fmt.Sprintf("%s/thumb?%s", g.BaseURL, uv.Encode())
}

func isPictureExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".gif", ".png":
		return true
	}
	return false
}
