package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//serveImage serves an image on an HTML page
func serveImage(w http.ResponseWriter, r *http.Request) {

	//Create a gallery object for creating links
	g := new(Gallery)
	g.BaseURL = cfg.BaseURL
	g.URLPath = strings.Trim(r.FormValue("gallery"), `/\`)
	g.GalleryPath = filepath.Join(filepath.Dir(os.Args[0]), galleryPath, g.URLPath)

	//Parse config
	err := g.ReadConfig()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	g.PageSize = g.NumImgRows * numCols

	//Create ImagePage object
	ip := new(ImagePage)
	ip.BgColor = g.BgColor
	ip.NavColor = g.NavColor
	ip.ShadowColor = g.ShadowColor
	ip.HeaderIMG = g.HeaderIMG
	ip.BaseURL = g.BaseURL
	ip.Gallery = g.URLPath
	ip.Page = r.FormValue("page")
	ip.FileName = strings.Trim(r.FormValue("name"), `./\`)
	ip.GalleryURL = fmt.Sprintf("%s/%s", g.BaseURL, g.URLPath)
	ip.ToPageURL = fmt.Sprintf("%s?page=%s", ip.GalleryURL, ip.Page)

	//Non-zero based page for display
	ip.PagePlusOne, _ = strconv.Atoi(ip.Page)
	ip.PagePlusOne++

	//Get URL to current image and header
	g.ServeImageRAW = true
	ip.HeaderFileURL = g.FileURL(fileTypeGlobal, ip.HeaderIMG, false)
	ip.ImageURL = g.FileURL(fileTypeGallery, ip.FileName, false)

	//Get gallery URL to previous and next image
	g.ServeImageRAW = false
	previmg, nextimg := getPrevAndNextFile(filepath.Join(g.GalleryPath, galleryImgPath), ip.FileName)
	g.Page = int(previmg.Num / g.PageSize)
	ip.PrevImageURL = g.FileURL(fileTypeGallery, previmg.Name, false)
	g.Page = int(nextimg.Num / g.PageSize)
	ip.NextImageURL = g.FileURL(fileTypeGallery, nextimg.Name, false)

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
	ShadowColor   string
	HeaderIMG     string
	BaseURL       string
	FileName      string
	Gallery       string
	Page          string
	PagePlusOne   int
	GalleryURL    string
	ToPageURL     string
	HeaderFileURL string
	ImageURL      string
	NextImageURL  string
	PrevImageURL  string
}

type ImgFile struct {
	Name string //filename
	Num  int    //filenumber in dir (to determine page it is on)
}

//get the URL to the previous and next image of img
func getPrevAndNextFile(startpath, img string) (ImgFile, ImgFile) {
	prev, next := ImgFile{}, ImgFile{}
	found := false
	i := 0
	filepath.Walk(startpath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && path != startpath {
			return filepath.SkipDir //skip any subdir
		}
		n := info.Name()
		if !isPictureExt(filepath.Ext(n)) {
			return nil
		}
		i++
		if n == img {
			found = true //found file img
			return nil
		} else if !found {
			prev.Name = n
			prev.Num = i - 1
		}
		if found {
			next.Name = n
			next.Num = i - 1
			return filepath.SkipDir //skip the rest of files
		}
		return nil
	})
	return prev, next
}
