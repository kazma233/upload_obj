package main

import (
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type FileLoader struct {
	http.Handler
}

func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "upload_obj",
		Width:  768,
		Height: 1024,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: NewFileLoader(),
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func (h *FileLoader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sp := strings.Split(r.URL.Path, "static/")
	if len(sp) < 2 {
		fmt.Println("URL path: ", r.URL.Path)
		return
	}

	// 解码URL编码的路径
	decodedBs, err := url.QueryUnescape(sp[1])
	if err != nil {
		fmt.Printf("URL decode error: %v\n", err)
		return
	}

	decodedPath := string(decodedBs)

	// 打印原始解码路径
	fmt.Printf("URL decoded path: %s\n", decodedPath)

	// 将URL路径转换为本地文件系统路径
	// decodedPath = filepath.FromSlash(decodedPath)
	// decodedPath = filepath.Clean(decodedPath)

	// 确保路径是绝对路径
	if !filepath.IsAbs(decodedPath) {
		fmt.Printf("Path is not absolute: %s\n", decodedPath)
		return
	}

	// 打印最终路径
	fmt.Printf("Final file path: %s\n", decodedPath)

	// 检查文件是否存在
	if _, err := os.Stat(decodedPath); os.IsNotExist(err) {
		fmt.Printf("File does not exist: %s\n", decodedPath)
		return
	} else if err != nil {
		fmt.Printf("File stat error: %v\n", err)
		return
	}

	// 打开文件
	file, err := os.Open(decodedPath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("File stat error: %v\n", err)
		return
	}

	// 设置Content-Type
	ext := strings.ToLower(filepath.Ext(decodedPath))
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	}
	w.Header().Set("Content-Type", contentType)

	// 提供文件内容
	http.ServeContent(w, r, filepath.Base(decodedPath), fileInfo.ModTime(), file)
}
