package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	fileTypeGallery = 0 //only image extension
	fileTypeGlobal  = 1 //only image extension
	fileTypeVideo   = 2 //only video extension
)

//serveFile serves image or video files
func serveFile(w http.ResponseWriter, r *http.Request) {

	path := filepath.Join(filepath.Dir(os.Args[0]), galleryPath, r.FormValue("gallery"))
	file := strings.Trim(r.FormValue("name"), `./\`)
	filetype, _ := strconv.Atoi(r.FormValue("filetype"))
	thumb, _ := strconv.ParseBool(r.FormValue("thumb"))

	switch filetype {
	case fileTypeGlobal:
		//do nothing
	case fileTypeGallery:
		path = filepath.Join(path, galleryImgPath)
	case fileTypeVideo:
		path = filepath.Join(path, galleryVidPath)
	}
	if thumb {
		path = filepath.Join(path, thumbPath)
	}
	path = filepath.Join(path, file)
	//extra safety to serve only allowed filetypes
	ext := filepath.Ext(path)
	if (thumb || filetype == fileTypeGallery || filetype == fileTypeGlobal) && !isPictureExt(ext) {
		fmt.Fprintln(w, "Not an image file")
		return
	}
	if (!thumb && filetype == fileTypeVideo) && !isVideoExt(ext) {
		fmt.Fprintln(w, "Not a video file")
		return
	}

	//Or ioutil.ReadFile...
	fh, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(w, "Cannot open file %s", r.FormValue("file"))
		return
	}
	defer fh.Close()

	//serve file
	stat, err := fh.Stat()
	if err != nil {
		fmt.Fprintf(w, "Cannot open file %s", r.FormValue("file"))
		return
	}
	http.ServeContent(w, r, r.FormValue("file"), stat.ModTime(), fh)
	return
}

func isPictureExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".gif", ".png":
		return true
	}
	return false
}

func isVideoExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp4", ".webm", ".avi", ".mkv", ".mpeg", ".mpg":
		return true
	}
	return false
}
