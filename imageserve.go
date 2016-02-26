package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//serveImage serves an image on an HTML page
func serveImage(w http.ResponseWriter, r *http.Request) {

	//Create a gallery object for creating links
	g := new(Gallery)
	g.BaseURL = baseURL
	g.URLPath = strings.Trim(r.FormValue("gallery"), `/\`)
	g.GalleryPath = filepath.Join(filepath.Dir(os.Args[0]), galleryPath, g.URLPath)
	g.ServeImageRAW = true

	//Parse config
	err := g.ReadConfig()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	//Create ImagePage object
	ip := new(ImagePage)
	ip.BgColor = g.BgColor
	ip.NavColor = g.NavColor
	ip.HeaderIMG = g.HeaderIMG
	ip.BaseURL = baseURL
	ip.Gallery = g.URLPath
	ip.Page = r.FormValue("page")
	ip.FileName = strings.Trim(r.FormValue("name"), `./\`)
	ip.GalleryURL = fmt.Sprintf("%s/%s", baseURL, g.URLPath)
	ip.ToPageURL = fmt.Sprintf("%s?page=%s", ip.GalleryURL, ip.Page)
	ip.HeaderFileURL = g.FileURL(fileTypeGlobal, ip.HeaderIMG, "f")
	ip.ImageURL = g.FileURL(fileTypeGallery, ip.FileName, "f")

	//Parse template
	tmpl, err := template.New("image.html").ParseFiles(
		filepath.Join(filepath.Dir(os.Args[0]), templatePath, "image.html"),
	)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	//Render template
	err = tmpl.Execute(w, ip)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	return
}

type ImagePage struct {
	BgColor       string
	NavColor      string
	HeaderIMG     string
	BaseURL       string
	FileName      string
	Gallery       string
	Page          string
	GalleryURL    string
	ToPageURL     string
	HeaderFileURL string
	ImageURL      string
}
