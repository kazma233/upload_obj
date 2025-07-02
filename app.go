package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"upload_obj/internal/upload"
	"upload_obj/internal/watermark"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var tempDir string

// App struct
type App struct {
	ctx context.Context
}

type AppConfig struct {
	upload.Config
	Watermark *watermark.WatermarkHandle `json:"watermark"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// init temp dir
	tempDir = filepath.Join(os.TempDir(), "upload_obj")
	os.Mkdir(tempDir, 0755)
}

func (a *App) SelectFile() string {
	fp, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select a file",
	}) // Open a file dialog
	if err != nil {
		runtime.LogError(a.ctx, "Failed to open file dialog: "+err.Error())
		return ""
	}

	return fp
}

func (a *App) UploadFile(filePath string, strategyType upload.BedType) (string, error) {
	strategy := upload.FindStrategy(strategyType)
	if strategy == nil {
		runtime.LogError(a.ctx, "Failed to find strategy")
		return "", errors.New("failed to find strategy")
	}

	savePath, err := WatermarkByPath(filePath, false)
	if err != nil {
		runtime.LogError(a.ctx, "Failed to watermark by path: "+err.Error())
		return "", err
	}

	url, err := strategy.UploadByPath(savePath)
	if err != nil {
		runtime.LogError(a.ctx, "Failed to upload file: "+err.Error())
		return "", err
	}

	return url, nil
}

func (a *App) Preview(filePath string) (string, error) {
	savePath, err := WatermarkByPath(filePath, true)
	if err != nil {
		runtime.LogError(a.ctx, "Failed to watermark by path: "+err.Error())
		return "", err
	}

	return "/static/" + url.QueryEscape(savePath), nil
}

func WatermarkByPath(filePath string, random bool) (string, error) {
	configJson, err := loadConfig()
	if err != nil {
		return "", err
	}

	var config AppConfig
	err = json.Unmarshal([]byte(configJson), &config)
	if err != nil {
		return "", err
	}

	// no need watermark
	if config.Watermark == nil || !config.Watermark.Check() {
		return filePath, nil
	}

	bgImg, err := watermark.OpenImg(filePath)
	if err != nil {
		return "", err
	}

	img, err := config.Watermark.Do(bgImg)
	if err != nil {
		return "", err
	}

	var savePath string
	if random {
		savePath = filepath.Join(tempDir, filepath.Base(filePath)) + "." + strconv.FormatInt(time.Now().Unix(), 10) + ".png"
	} else {
		savePath = filepath.Join(tempDir, filepath.Base(filePath)) + ".watermark.png"
	}
	err = watermark.Save2PngImg(img, savePath)
	if err != nil {
		return "", err
	}

	return savePath, nil
}

func (a *App) SaveConfig(configJson string) (string, error) {
	// save configJson at ./config.json
	configFile, err := os.OpenFile("./config.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		runtime.LogError(a.ctx, "Failed to open file dialog: "+err.Error())
		return "", err
	}
	_, err = io.WriteString(configFile, configJson)
	if err != nil {
		runtime.LogError(a.ctx, "Failed to write file dialog: "+err.Error())
		return "", err
	}

	err = initUpload()
	if err != nil {
		runtime.LogError(a.ctx, "Failed to init upload: "+err.Error())
		return "", err
	}

	return "", nil
}

func (a *App) LoadConfig() (string, error) {
	configJson, err := loadConfig()
	if err != nil {
		runtime.LogError(a.ctx, "Failed to load config: "+err.Error())
		return "", err
	}

	err = initUpload()
	if err != nil {
		runtime.LogError(a.ctx, "Failed to init upload: "+err.Error())
		return "", err
	}

	return configJson, nil
}

func loadConfig() (string, error) {
	// load configJson at ./config.json
	configFile, err := os.OpenFile("./config.json", os.O_RDONLY, 0644)
	if err != nil {
		return "", err
	}
	configJson, err := io.ReadAll(configFile)
	if err != nil {
		return "", err
	}
	return string(configJson), nil
}

func initUpload() (err error) {
	configJson, err := loadConfig()
	if err != nil {
		return
	}
	err = upload.InitUpload(configJson)
	return
}
