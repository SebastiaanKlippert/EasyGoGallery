package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func serveFile(w http.ResponseWriter, r *http.Request) {

	path := ""
	size := r.FormValue("size")
	if size == "" {
		//original
		path = filepath.Join(filepath.Dir(os.Args[0]), galleryPath, r.FormValue("gallery"), r.FormValue("file"))
	} else {
		//thumbnail
		path = filepath.Join(filepath.Dir(os.Args[0]), galleryPath, r.FormValue("gallery"), fmt.Sprintf("thumbs%s", size), r.FormValue("file"))
	}

	//Or ioutil.ReadFile...
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Cannot open file %s", r.FormValue("file"))
		return
	}
	defer file.Close()
	//set content type
	mimeType := mime.TypeByExtension(filepath.Ext(r.FormValue("file")))
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	}
	//output data
	io.Copy(w, file)

	return
}
