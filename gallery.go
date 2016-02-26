package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	//TODO config
	baseURL        = "http://localhost"
	galleryPath    = "./galleries"
	galleryImgPath = "/img"
	galleryVidPath = "/vid"
	thumbPath      = "/thumbs"
	templatePath   = "./templates"
	configFile     = "config.json"
	numCols        = 12
	numRows        = 12
	numVidCols     = 3
	numVidRows     = 16
	imgPadding     = 10
	vidWidth       = 480
	vidHeight      = 360
	serveImageRAW  = false //serve large images as is or in an HTML page
)

type GalleryConfig struct {
	BgColor   string
	NavColor  string
	HeaderIMG string
}

func serveGallery(w http.ResponseWriter, r *http.Request) {

	//Create gallery object
	g := new(Gallery)
	g.BaseURL = baseURL
	g.URLPath = strings.TrimSuffix(strings.Trim(r.URL.Path, `/\`), "/movies")
	g.GalleryURL = fmt.Sprintf("%s/%s", baseURL, g.URLPath)
	g.VideoURL = fmt.Sprintf("%s%s", g.GalleryURL, "/movies")
	g.ServingVideo = strings.HasSuffix(r.URL.Path, "/movies")
	g.Page, _ = strconv.Atoi(r.FormValue("page"))
	g.PageSize = numRows * numCols
	g.GalleryPath = filepath.Join(filepath.Dir(os.Args[0]), galleryPath, g.URLPath)
	g.VidWidth = vidWidth
	g.VidHeight = vidHeight
	g.ServeImageRAW = serveImageRAW

	//Check if image gallery exists
	_, err := os.Stat(filepath.Join(g.GalleryPath, galleryImgPath))
	if err != nil {
		fmt.Fprintf(w, "Cannot find gallery %q\n", g.URLPath)
		return
	}
	//Check if image thumb dir exists
	//TODO Create it
	_, err = os.Stat(filepath.Join(g.GalleryPath, galleryImgPath, thumbPath))
	if err != nil {
		fmt.Fprintf(w, "Cannot find image thumb dir %q\n", filepath.Join(g.GalleryPath, galleryImgPath, thumbPath))
		return
	}
	//Check if there are videos
	_, err = os.Stat(filepath.Join(g.GalleryPath, galleryVidPath))
	if err == nil {
		g.HasVideos = true
	}
	if g.HasVideos {
		_, err = os.Stat(filepath.Join(g.GalleryPath, galleryVidPath, thumbPath))
		if err != nil {
			fmt.Fprintf(w, "Cannot find video thumb dir %q\n", filepath.Join(g.GalleryPath, galleryVidPath, thumbPath))
			return
		}
	}

	//Parse config
	err = g.ReadConfig()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	//Fill number of columns
	numC := numCols
	if g.ServingVideo {
		numC = numVidCols
	}
	for i := 0; i < numC; i++ {
		g.Columns = append(g.Columns, i)
	}

	//Fill number of rows
	numR := numRows
	if g.ServingVideo {
		numR = numVidRows
	}
	for i := 0; i < numR; i++ {
		g.Rows = append(g.Rows, i)
	}

	//List thumbs, fow now we assume there is a thumb present for each image
	g.listThumbs(g.ServingVideo)

	//Add template functions
	fm := template.FuncMap{
		"FileURL":      g.FileURL,
		"ImgURLPos":    g.ImgURLPos,
		"ImgPosExists": g.ImgPosExists,
		"PageURL":      g.PageURL,
		"IsLastPage":   g.IsLastPage,
	}

	//Parse template
	base := filepath.Join(filepath.Dir(os.Args[0]), templatePath)
	tmpl, err := template.New("gallery.html").Funcs(fm).ParseFiles(
		filepath.Join(base, "gallery.html"),
		filepath.Join(base, "images.html"),
		filepath.Join(base, "movies.html"),
	)
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
	NavColor  string
	HeaderIMG string

	BaseURL string
	URLPath string

	GalleryURL string
	VideoURL   string

	HasVideos    bool
	ServingVideo bool

	Page     int
	PageSize int

	VidWidth  int
	VidHeight int

	Columns []int
	Rows    []int

	GalleryPath string

	FileNames []string

	ServeImageRAW bool

	skipImages, skippedImages int
	walkPath                  string
}

func (g *Gallery) listThumbs(videos bool) {
	if videos {
		g.walkPath = filepath.Join(g.GalleryPath, galleryVidPath, thumbPath)
	} else {
		g.walkPath = filepath.Join(g.GalleryPath, galleryImgPath, thumbPath)
	}
	//Set number of images to skip
	g.skipImages = g.Page * g.PageSize
	g.skippedImages = 0

	g.Error = filepath.Walk(g.walkPath, g.collectImages)
}

func (g *Gallery) collectImages(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() && path != g.walkPath { //skip all subdirectories
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
	if len(g.FileNames) >= g.PageSize+1 { //+1 to determine if last page reached
		return filepath.SkipDir
	}
	return nil
}

//FileURL returns the GET URL for file, if file is a thumb, thumb should be set to "t"
//The filetype is 0 for gallery images, 1 for global images or 2 for videos
//The returned linked is in format http://localhost/file?name=xxx.ext&filetype=1&gallery=xxx&thumb=true for raw files or
//http://localhost/image?name=xxx.ext&filetype=0&gallery=xxx for gallery image pages
//Where http://localhost/ is the baseURL.
func (g *Gallery) FileURL(filetype int, file, thumb string) string {
	uv := make(url.Values)
	if filetype == fileTypeVideo && thumb != "t" {
		//strip one extension (thumb is xxx.mp4.jpg, video is xxx.mp4)
		file = strings.TrimSuffix(file, filepath.Ext(file))
	}
	how := "file"
	if thumb != "t" && filetype == fileTypeGallery && g.ServeImageRAW == false {
		how = "image"
		uv.Add("page", strconv.Itoa(g.Page))
	} else {
		uv.Add("filetype", strconv.Itoa(filetype))
	}
	uv.Add("gallery", g.URLPath)
	uv.Add("name", file)
	if thumb == "t" {
		uv.Add("thumb", "true")
	}
	return fmt.Sprintf("%s/%s?%s", g.BaseURL, how, uv.Encode())
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
	filetype := fileTypeGallery
	if g.ServingVideo {
		filetype = fileTypeVideo
	}
	return g.FileURL(filetype, g.FileNames[p], thumb)
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

//IsLastPage returns if the last page has been reached
func (g *Gallery) IsLastPage() bool {
	return len(g.FileNames) > g.PageSize
}

//ReadConfig reads config.json for this gallery from disk and sets the values in g
func (g *Gallery) ReadConfig() error {
	jsonb, err := ioutil.ReadFile(filepath.Join(g.GalleryPath, configFile))
	if err != nil {
		return err
	}
	config := new(GalleryConfig)
	err = json.Unmarshal(jsonb, config)
	if err != nil {
		return err
	}
	g.BgColor = config.BgColor
	g.NavColor = config.NavColor
	g.HeaderIMG = config.HeaderIMG
	return nil
}
