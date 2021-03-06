package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

type filelistResponse struct {
	Files []string `json:"files"`
}

func filelist() {
	http.HandleFunc("/api/filelist/", func(rw http.ResponseWriter, rq *http.Request) {
		if !(rq.Method == "GET" || rq.Method == "") {
			return
		}

		resStruct := filelistResponse{}

		files, err := ioutil.ReadDir("./files")
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			resStruct.Files = append(resStruct.Files, f.Name())
		}

		resBytes, err := json.Marshal(resStruct)

		if err != nil {
			log.Fatal(err)
		}

		rw.Write(resBytes)
	})
}

func files() {
	http.Handle("/api/files/", http.StripPrefix("/api/files/", http.FileServer(http.Dir("./files"))))
}

func upload() {
	http.HandleFunc("/api/upload", func(rw http.ResponseWriter, rq *http.Request) {
		// maximum upload of 100 MB files
		rq.ParseMultipartForm(100 << 20)

		// "files" from the html file input's `name` attribute
		fhs := rq.MultipartForm.File["files"]
		for _, fh := range fhs {
			src, err := fh.Open()
			if err != nil {
				log.Fatal(err)
			}

			dst, err := os.OpenFile("./files/"+fh.Filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
			if err != nil {
				panic(err)
			}
			defer dst.Close()

			_, err = io.Copy(dst, src)
			if err != nil {
				log.Fatal(err)
			}
		}

		http.Redirect(rw, rq, "/", http.StatusSeeOther)
	})
}

func clear() {
	http.HandleFunc("/api/clear", func(rw http.ResponseWriter, rq *http.Request) {
		if rq.Method != "POST" {
			return
		}
		files, err := ioutil.ReadDir("./files")
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			err = os.Remove("./files/" + f.Name())
			if err != nil {
				log.Fatal(err)
			}
		}

	})
}

func statics() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
}

func createDirIfNotExists() {
	err := os.MkdirAll("./files", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

// Get preferred outbound ip of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func main() {
	createDirIfNotExists()
	filelist()
	files()
	upload()
	clear()
	statics()

	port := 8080
	for {
		listener, err := net.Listen("tcp", ":"+fmt.Sprint(port))
		if err != nil {
			port++
			continue
		}
		fmt.Printf("Listening on %s:%d\n", getOutboundIP(), listener.Addr().(*net.TCPAddr).Port)
		log.Fatal(http.Serve(listener, nil))
	}
}
