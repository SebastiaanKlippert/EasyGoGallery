package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

func index(w http.ResponseWriter, r *http.Request) {

	idx := new(Index)
	idx.Title = "EasyGoGallery"
	idx.Galleries = listGalleries(filepath.Join(filepath.Dir(os.Args[0]), galleryPath))

	//Parse template
	tmpl, err := template.New("index.html").ParseFiles(
		filepath.Join(filepath.Dir(os.Args[0]), templatePath, "index.html"),
	)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	//Render template
	err = tmpl.Execute(w, idx)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	return
}

type Index struct {
	Title     string
	Galleries []ListGallery
}

type ListGallery struct {
	Name   string
	URL    string
	ImgURL string
}

//Each dir in /galleries that has a config.json file in it that can
//be parsed as our galery config qualifies as a gallery
func listGalleries(startpath string) []ListGallery {

	galleries := []ListGallery{}

	filepath.Walk(startpath, func(path string, info os.FileInfo, wErr error) error {
		if wErr != nil {
			return wErr
		}
		//get relative dir from startdir, we only go one level deep, each subdir in
		//./galleries can have a config file, we do not need to travel through more subdirs
		if info.IsDir() {
			rel, err := filepath.Rel(startpath, path)
			if err != nil {
				return err
			}
			if rel != filepath.Base(rel) {
				return filepath.SkipDir
			}
		}
		if info.Name() != configFile {
			return nil
		}
		//We are on a gallery config file, see if we can parse it
		g := new(Gallery)
		g.GalleryPath = filepath.Dir(path)
		perr := g.ReadConfig()
		if perr != nil {
			return nil
		}
		//No parse error, this is a gallery
		g.BaseURL = cfg.BaseURL
		g.URLPath = filepath.Base(g.GalleryPath)

		lg := ListGallery{}
		lg.Name = filepath.Base(g.GalleryPath)
		lg.URL = fmt.Sprintf("%s/%s", g.BaseURL, filepath.Base(g.GalleryPath))
		lg.ImgURL = g.FileURL(fileTypeGlobal, g.HeaderIMG, false)

		galleries = append(galleries, lg)
		return nil
	})
	return galleries
}
