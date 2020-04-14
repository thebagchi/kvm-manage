package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	MDCStyleSheet = "https://cdnjs.cloudflare.com/ajax/libs/material-components-web/5.1.0/material-components-web.min.css"
	MDCJavascript = "https://cdnjs.cloudflare.com/ajax/libs/material-components-web/5.1.0/material-components-web.min.js"
)

func DownloadFile(filepath string, url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, response.Body)
	return err
}

func CopyFile(src string, dst string) {
	data, err := ioutil.ReadFile(src)
	err = ioutil.WriteFile(dst, data, 0644)
	_ = err
}

func ParseEnv(cmd string) map[string]string {
	dump, err := exec.Command(cmd, strings.Fields("env")...).CombinedOutput()
	if nil != err {
		fmt.Println("Error: ", err)
		fmt.Println("Log: ", string(dump))
		return nil
	}
	env := map[string]string{
		// Empty
	}
	items := strings.Split(string(dump), "\n")
	for _, item := range items {
		item = strings.TrimSpace(item)
		items := strings.Split(item, "=")
		if len(items) == 2 {
			var (
				key   = items[0]
				value = items[1]
			)
			env[key] = strings.TrimPrefix(strings.TrimSuffix(value, "\""), "\"")
		}
	}
	return env
}

func Build() {
	env := ParseEnv("go")
	if nil == env || len(env) == 0 {
		fmt.Println("Failed executing command \"go env\"")
		return
	}
	var (
		goos   = os.Getenv("GOOS")
		goarch = os.Getenv("GOARCH")
		root   = env["GOROOT"]
	)
	_ = os.Setenv("GOOS", "js")
	_ = os.Setenv("GOARCH", "wasm")
	defer func() {
		_ = os.Setenv("GOOS", goos)
		_ = os.Setenv("GOARCH", goarch)
	}()
	dump, err := exec.Command(
		"go",
		strings.Fields("build -o www/main.wasm frontend/main.go")...,
	).CombinedOutput()
	if nil != err {
		fmt.Println("Error: ", err)
		fmt.Println("Log: ", string(dump))
		return
	}
	CopyFile(filepath.Join(root, "/misc/wasm/wasm_exec.js"), "www/wasm_exec.js")
	CopyFile("frontend/index.html", "www/index.html")
	{
		items := strings.Split(MDCStyleSheet, "/")
		err := DownloadFile(filepath.Join("www", items[len(items)-1]), MDCStyleSheet)
		if nil != err {
			fmt.Println("Error: ", err)
		}
	}
	{
		items := strings.Split(MDCJavascript, "/")
		err := DownloadFile(filepath.Join("www", items[len(items)-1]), MDCJavascript)
		if nil != err {
			fmt.Println("Error: ", err)
		}
	}
}

func main() {
	Build()
}
