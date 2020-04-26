package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

var mux map[string]func(http.ResponseWriter, *http.Request)

type _FileServiceHandler struct{}
type _Home struct {
	Title string
}

const (
	strTemplateDir = "./view/"
	strUploadDir   = "./upload/"
)

func main() {

	if _, err := os.Stat(strUploadDir); os.IsNotExist(err) {
		os.Mkdir(strUploadDir, 0777)
		os.Chmod(strUploadDir, 0777)
	}

	server := http.Server{
		Addr:        ":9090",
		Handler:     &_FileServiceHandler{},
		ReadTimeout: 10 * time.Second,
	}
	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	mux["/"] = index
	mux["/upload"] = upload
	mux["/file"] = _StaticServer
	server.ListenAndServe()
}

func (*_FileServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.String()]; ok {
		h(w, r)
		return
	}
	if ok, _ := regexp.MatchString("/css/", r.URL.String()); ok {
		http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))).ServeHTTP(w, r)
	} else {
		http.StripPrefix("/", http.FileServer(http.Dir("./upload/"))).ServeHTTP(w, r)
	}

}

func upload(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t, _ := template.ParseFiles(strTemplateDir + "file.html")
		t.Execute(w, "上传文件")
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Fprintf(w, "%v", "上传错误")
			return
		}
		fileext := filepath.Ext(handler.Filename)
		if check(fileext) == false {
			fmt.Fprintf(w, "%v", "不允许的上传类型")
			return
		}
		filename := strconv.FormatInt(time.Now().Unix(), 10) + fileext
		f, _ := os.OpenFile(strUploadDir+filename, os.O_CREATE|os.O_WRONLY, 0660)
		_, err = io.Copy(f, file)
		if err != nil {
			fmt.Fprintf(w, "%v", "上传失败")
			return
		}
		filedir, _ := filepath.Abs(strUploadDir + filename)
		fmt.Fprintf(w, "%v", filename+"上传完成,服务器地址:"+filedir)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	title := _Home{Title: "首页"}
	t, _ := template.ParseFiles(strTemplateDir + "index.html")
	t.Execute(w, title)
}

func _StaticServer(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/file", http.FileServer(http.Dir("./upload/"))).ServeHTTP(w, r)
}

func check(name string) bool {
	ext := []string{".exe", ".js", ".png"}

	for _, v := range ext {
		if v == name {
			return false
		}
	}
	return true
}
