package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
        "strings"
)

func serveFile(w http.ResponseWriter, r *http.Request) {

	path := ""
        file := strings.Trim(r.FormValue("file"),  `./\`)
        size := r.FormValue("size")
	if size == "" {
		//original
		path = filepath.Join(filepath.Dir(os.Args[0]), galleryPath, r.FormValue("gallery"), file)
	} else {
		//thumbnail
		path = filepath.Join(filepath.Dir(os.Args[0]), galleryPath, r.FormValue("gallery"), fmt.Sprintf("thumbs%s", size), file)
	}

        ext := filepath.Ext(path)
        if !isPictureExt(ext) {
           fmt.Fprintln(w, "Not an image file")
           return
        } 

	//Or ioutil.ReadFile...
	fh, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Cannot open file %s", r.FormValue("file"))
		return
	}
	defer fh.Close()
	//set content type
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	}
	//output data
	io.Copy(w, fh)

	return
}
