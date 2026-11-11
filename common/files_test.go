package common

import (
	"bytes"
	"io/fs"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sohaha/zlsgo/znet"
)

type uploadPart struct {
	name string
	data []byte
}

func newUploadContext(t *testing.T, parts ...uploadPart) *znet.Context {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	for _, p := range parts {
		fw, err := w.CreateFormFile("file", p.name)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := fw.Write(p.data); err != nil {
			t.Fatalf("write form file: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return znet.New().NewContext(httptest.NewRecorder(), req)
}

func TestUploadKeepsEachFileBody(t *testing.T) {
	oldDir := Define.UploadLocalDir
	Define.UploadLocalDir = t.TempDir()
	t.Cleanup(func() { Define.UploadLocalDir = oldDir })

	first := []byte("first-file-payload-1234")
	second := []byte("two")

	results, err := Upload(newUploadContext(t,
		uploadPart{name: "a.txt", data: first},
		uploadPart{name: "b.txt", data: second},
	), "docs")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}
	if !bytes.Equal(results[0].Body, first) {
		t.Fatalf("first body corrupted: %q", results[0].Body)
	}
	if !bytes.Equal(results[1].Body, second) {
		t.Fatalf("second body corrupted: %q", results[1].Body)
	}
	if results[0].Path == results[1].Path {
		t.Fatal("distinct files must have distinct paths")
	}

	saved := map[string]bool{}
	err = filepath.WalkDir(Define.UploadLocalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		saved[string(data)] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walk upload dir: %v", err)
	}
	if !saved[string(first)] || !saved[string(second)] {
		t.Fatalf("saved files do not match uploaded bodies: %v", saved)
	}
}

func TestUploadRejectsOversizedFile(t *testing.T) {
	oldDir := Define.UploadLocalDir
	Define.UploadLocalDir = t.TempDir()
	t.Cleanup(func() { Define.UploadLocalDir = oldDir })

	_, err := Upload(newUploadContext(t,
		uploadPart{name: "big.txt", data: []byte("0123456789")},
	), "", func(o *UploadOption) { o.MaxSize = 4 })
	if err == nil {
		t.Fatal("expected oversized upload to fail")
	}
}
