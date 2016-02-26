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
	baseURL      = "http://localhost"
	galleryPath  = "./galleries"
	templatePath = "./templates"
	thumbSize    = 120
	numCols      = 12
	numRows      = 20
	imgPadding   = 10
	//TODO config per gallery
	bgColor   = "#faf8ec"
	headerIMG = "header.png"
)

func serveGallery(w http.ResponseWriter, r *http.Request) {

	//Create gallery object
	g := new(Gallery)
	g.BgColor = bgColor
	g.HeaderIMG = headerIMG
	g.BaseURL = baseURL
	g.URLPath = strings.Trim(r.URL.Path, `/\`)
	g.Page, _ = strconv.Atoi(r.FormValue("page"))
	g.PageSize = numRows * numCols
	g.ThumbSize = thumbSize
	g.ImagePath = filepath.Join(filepath.Dir(os.Args[0]), galleryPath, g.URLPath)
	g.ThumbPath = filepath.Join(g.ImagePath, fmt.Sprintf("thumbs%d", g.ThumbSize))

	//Check if gallery exists
	_, err := os.Stat(g.ImagePath)
	if err != nil {
		fmt.Fprintf(w, "Cannot find gallery %q\n", g.URLPath)
		return
	}
	//Check if thumb dir exists
	//TODO Create it
	_, err = os.Stat(g.ThumbPath)
	if err != nil {
		fmt.Fprintf(w, "Cannot find ThumbPath %q\n", g.ThumbPath)
		return
	}

	//Fill number of columns
	for i := 0; i < numCols; i++ {
		g.Columns = append(g.Columns, i)
	}

	//Fill number of rows
	for i := 0; i < numRows; i++ {
		g.Rows = append(g.Rows, i)
	}

	//List images, fow now we assume there is a thumb present for each image
	//TODO list images and create missing thumbs
	g.listImages(false)

	//Add template functions
	fm := template.FuncMap{
		"ImgURL":       g.ImgURL,
		"ImgURLPos":    g.ImgURLPos,
		"ImgPosExists": g.ImgPosExists,
		"PageURL":      g.PageURL,
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

	BgColor   string
	HeaderIMG string

	BaseURL string
	URLPath string

	Page     int
	PageSize int

	Columns []int
	Rows    []int

	ImagePath string
	ThumbSize int
	ThumbPath string

	FileNames []string

	skipImages, skippedImages int
}

func (g *Gallery) listImages(thumbs bool) {
	path := g.ImagePath
	if thumbs {
		path = g.ThumbPath
	}
	//Set number of images to skip
	g.skipImages = g.Page * g.PageSize
	g.skippedImages = 0

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
	if g.skippedImages < g.skipImages {
		g.skippedImages++
		return nil
	}
	g.FileNames = append(g.FileNames, name)
	if len(g.FileNames) >= g.PageSize {
		return filepath.SkipDir
	}
	return nil
}

//ImgURL returns the GET URL for image img, if img is a thumb, thumb should be set to "t"
//The returned linked is in format http://localhost/image?file=xxx.ext&gallery=xxx
//Where http://localhost/ is the baseURL.
func (g *Gallery) ImgURL(img, thumb string) string {
	uv := make(url.Values)
	uv.Add("gallery", g.URLPath)
	uv.Add("file", img)
	if thumb == "t" {
		uv.Add("size", strconv.Itoa(g.ThumbSize))
	}
	return fmt.Sprintf("%s/image?%s", g.BaseURL, uv.Encode())
}

//ImgURLPos returns the GET URL ImgURL for the image at row r and col c at g.Page,
//if img is a thumb, thumb should be set to "t".
//Both r and c are zero-based.
func (g *Gallery) ImgURLPos(r, c int, thumb string) string {
	//calculate image pos
	p := (r * len(g.Columns)) + c
	if p < 0 || len(g.FileNames)-1 < p {
		return ""
	}
	return g.ImgURL(g.FileNames[p], thumb)
}

//ImgPosExists returns if there is an image present at row r and col c at g.Page
func (g *Gallery) ImgPosExists(r, c int) bool {
	//calculate image pos
	p := (r * len(g.Columns)) + c
	if p < 0 || len(g.FileNames)-1 < p {
		return false
	}
	return true
}

//PageURL returns the URL to the next (if next == "t") or previous page
func (g *Gallery) PageURL(next string) string {
	p := g.Page
	if next == "t" {
		p++
	} else {
		p--
	}
	return fmt.Sprintf("%s/%s?page=%d", baseURL, g.URLPath, p)
}

func isPictureExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".gif", ".png":
		return true
	}
	return false
}
