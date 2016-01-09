package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

func main() {
	args := os.Args[1:]
	if len(args) == 1 {
		log.Fatal(serve(args[0]))
	} else if len(args) >= 2 {
		if err := run(args...); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println(`usage:
colorizer address
colorizer input.jpg output.jpg [converted-input.jpg]
`)
	}
}

var colorizeTemplate = template.Must(template.New("").Parse(string(colorizeHTML)))

func serve(addr string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := colorize(w, r)
		if err != nil {
			serveError(w, err)
			return
		}
	})
	log.Println("listening on", addr)
	return http.ListenAndServe(addr, nil)
}

func run(args ...string) error {
	args = append([]string{"/colorize/colorize.py"}, args...)
	b, err := exec.Command("python", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, b)
	}
	return nil
}

func serveError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func colorize(w http.ResponseWriter, r *http.Request) error {
	var files [][]string
	if err := r.ParseMultipartForm(32 << 20); err == nil {
		for _, fhs := range r.MultipartForm.File {
			for _, fh := range fhs {
				dir, err := ioutil.TempDir("", "")
				defer os.RemoveAll(dir)
				if err != nil {
					return err
				}
				ifile, err := os.Create(filepath.Join(dir, "input"))
				if err != nil {
					return err
				}
				defer ifile.Close()
				f, err := fh.Open()
				if err != nil {
					return err
				}
				defer f.Close()
				if _, err := io.Copy(ifile, f); err != nil {
					return err
				}
				ofile := filepath.Join(dir, "output.jpg")
				tfile := filepath.Join(dir, "temp.jpg")
				if err = run(ifile.Name(), ofile, tfile); err != nil {
					return err
				}
				var b64 []string
				for _, name := range []string{tfile, ofile} {
					b, err := ioutil.ReadFile(name)
					if err != nil {
						return err
					}
					s := base64.StdEncoding.EncodeToString(b)
					b64 = append(b64, s)
				}
				files = append(files, b64)
			}
		}
	}
	return colorizeTemplate.Execute(w, files)
}

var colorizeHTML = []byte(`<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>colorizer</title>
	</head>
	<body>
		{{range .}}
		<div>
			{{range .}}
				<img alt="" src="data:image/jpg;base64,{{.}}" />
			{{end}}
		</div>
		{{end}}
		<p>Select images to upload:</p>
		<form enctype="multipart/form-data" method="POST">
			<input type="file" name="input" multiple>
			<p><input type="submit" value="colorize">
		</form>
	</body>
</html>
`)
