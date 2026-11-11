package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUploadRejectsDirectoryTraversal(t *testing.T) {
	root := t.TempDir()
	oldDir := Define.UploadLocalDir
	Define.UploadLocalDir = filepath.Join(root, "static")
	t.Cleanup(func() { Define.UploadLocalDir = oldDir })

	outside := filepath.Join(root, "escaped")

	cases := []struct {
		name   string
		subDir string
		dir    string
	}{
		{"subdir traversal", "../escaped", ""},
		{"nested subdir traversal", "docs/../../escaped", ""},
		{"dir traversal", "docs", "../../escaped"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := []func(o *UploadOption){}
			if tc.dir != "" {
				d := tc.dir
				opts = append(opts, func(o *UploadOption) { o.Dir = d })
			}
			_, err := Upload(newUploadContext(t, uploadPart{name: "a.txt", data: []byte("hi")}), tc.subDir, opts...)
			if err == nil {
				t.Fatalf("expected traversal to be rejected")
			}
			if _, statErr := os.Stat(outside); statErr == nil {
				t.Fatalf("upload escaped the upload root: %s exists", outside)
			}
		})
	}
}

func TestUploadDirOptionStaysUnderRoot(t *testing.T) {
	root := t.TempDir()
	oldDir := Define.UploadLocalDir
	Define.UploadLocalDir = filepath.Join(root, "static")
	t.Cleanup(func() { Define.UploadLocalDir = oldDir })

	results, err := Upload(newUploadContext(t, uploadPart{name: "a.txt", data: []byte("hi")}),
		"docs", func(o *UploadOption) { o.Dir = "images" })
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}

	found := ""
	_ = filepath.WalkDir(Define.UploadLocalDir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".txt") {
			found = p
		}
		return nil
	})
	if found == "" {
		t.Fatal("uploaded file not found")
	}
	wantDir := filepath.Join(Define.UploadLocalDir, "docs", "images")
	if filepath.Dir(found) != wantDir {
		t.Fatalf("saved in %s, want %s", filepath.Dir(found), wantDir)
	}
}

func TestUploadRelDir(t *testing.T) {
	cases := []struct {
		subDir  string
		dir     string
		want    string
		wantErr bool
	}{
		{"", "", "", false},
		{"docs", "", "docs", false},
		{"/docs/", "", "docs", false},
		{"a::b", "", "a__b", false},
		{"docs", "images", "docs/images", false},
		{"docs", "/images", "images", false},
		{"docs/./x//y", "", "docs/x/y", false},
		{"docs/../x", "", "", true},
		{"..", "", "", true},
		{"docs", "../x", "", true},
		{"docs", `..\x`, "", true},
	}
	for _, tc := range cases {
		got, err := uploadRelDir(tc.subDir, tc.dir)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("uploadRelDir(%q, %q): expected error, got %q", tc.subDir, tc.dir, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("uploadRelDir(%q, %q): %v", tc.subDir, tc.dir, err)
		}
		if got != tc.want {
			t.Fatalf("uploadRelDir(%q, %q) = %q, want %q", tc.subDir, tc.dir, got, tc.want)
		}
	}
}

func TestUploadRootRelativeDirIgnoresSubDir(t *testing.T) {
	root := t.TempDir()
	oldDir := Define.UploadLocalDir
	Define.UploadLocalDir = filepath.Join(root, "static")
	t.Cleanup(func() { Define.UploadLocalDir = oldDir })

	if _, err := Upload(newUploadContext(t, uploadPart{name: "a.txt", data: []byte("hi")}),
		"docs", func(o *UploadOption) { o.Dir = "/images" }); err != nil {
		t.Fatalf("upload: %v", err)
	}

	found := ""
	_ = filepath.WalkDir(Define.UploadLocalDir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".txt") {
			found = p
		}
		return nil
	})
	if found == "" {
		t.Fatal("uploaded file not found")
	}
	wantDir := filepath.Join(Define.UploadLocalDir, "images")
	if filepath.Dir(found) != wantDir {
		t.Fatalf("saved in %s, want %s", filepath.Dir(found), wantDir)
	}
}
