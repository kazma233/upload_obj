package upload

import (
	"io"
	"os"
	"testing"
)

func Test_upload(t *testing.T) {
	f, err := os.OpenFile("../../config.json", os.O_RDONLY, 0644)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()
	fb, err := io.ReadAll(f)
	if err != nil {
		t.Error(err)
		return
	}
	InitUpload(string(fb))
	p, err := FindStrategy(BedType(GITHUB)).UploadByPath(`./test/demo.jpg`)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(p)
}
