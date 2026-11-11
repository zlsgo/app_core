package common

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zstring"
)

type UploadOption struct {
	Key              string
	Dir              string
	MimeType         []string
	MaxSize          int64
	CustomFilter     func(r *UploadResult) error
	CustomProcessing func(r *UploadResult, dir string) error
}

type UploadResult struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Ext      string `json:"ext"`
	MimeType string `json:"mime_type"`
	Storage  string `json:"storage"`
	Size     int64  `json:"size"`
	Body     []byte `json:"-"`
}

// uploadRelDir builds the directory path, relative to the upload root, where
// files are written. subDirName and dir are treated as relative names: leading
// separators are ignored and "." segments are dropped. ".." segments are
// rejected so a value taken from a request can never escape the upload root.
// A dir that starts with "/" is resolved from the upload root and ignores
// subDirName, which keeps the historical behaviour of that option.
func uploadRelDir(subDirName, dir string) (string, error) {
	rel := strings.ReplaceAll(subDirName, "::", "__")
	if dir != "" {
		if strings.HasPrefix(dir, "/") {
			rel = dir
		} else {
			rel += "/" + dir
		}
	}

	segments := strings.Split(strings.ReplaceAll(rel, "\\", "/"), "/")
	cleaned := make([]string, 0, len(segments))
	for _, segment := range segments {
		switch segment {
		case "", ".":
			continue
		case "..":
			return "", zerror.InvalidInput.Text("上传目录不合法")
		}
		if strings.ContainsRune(segment, 0) {
			return "", zerror.InvalidInput.Text("上传目录不合法")
		}
		cleaned = append(cleaned, segment)
	}

	return strings.Join(cleaned, "/"), nil
}

func Upload(c *znet.Context, subDirName string, opt ...func(o *UploadOption)) ([]UploadResult, error) {
	o := UploadOption{
		Key:     "file",
		MaxSize: 1024 * 1024 * 2,
	}
	for _, p := range opt {
		p(&o)
	}
	files, err := c.FormFiles(o.Key)
	if err != nil {
		return nil, zerror.InvalidInput.Wrap(err, "上传失败")
	}

	relDir, err := uploadRelDir(subDirName, o.Dir)
	if err != nil {
		return nil, err
	}
	uploadDir := zfile.RealPathMkdir(filepath.Join(Define.UploadLocalDir, relDir), true)

	uploads := make([]UploadResult, 0, len(files))

	for _, v := range files {
		if v.Size > o.MaxSize {
			return nil, zerror.InvalidInput.Text("文件大小超过限制")
		}

		f, err := v.Open()
		if err != nil {
			return nil, zerror.InvalidInput.Wrap(err, "文件读取失败")
		}

		// Each file needs its own buffer: reusing one and calling Reset would
		// let the next file overwrite the previously collected Body slices.
		var buf bytes.Buffer
		_, copyErr := io.Copy(&buf, io.LimitReader(f, o.MaxSize+1))
		_ = f.Close()
		if copyErr != nil {
			return nil, zerror.InvalidInput.Wrap(copyErr, "文件读取失败")
		}

		b := buf.Bytes()

		if int64(len(b)) > o.MaxSize {
			return nil, zerror.InvalidInput.Text("文件大小超过限制")
		}

		mt := zfile.GetMimeType(v.Filename, b)
		n := strings.Split(mt, "/")
		if len(n) < 2 {
			return nil, zerror.InvalidInput.Text("文件类型无法识别")
		}

		mt = strings.SplitN(mt, ";", 2)[0]

		if len(o.MimeType) > 0 {
			ok := false
			for _, v := range o.MimeType {
				if v == mt || v == n[1] || zstring.Match(mt, v) {
					ok = true
					break
				}
			}

			if !ok {
				return nil, zerror.InvalidInput.Text("不支持的文件类型")
			}
		}

		ext := filepath.Ext(v.Filename)
		if ext == "" {
			if len(n) > 1 {
				ext = "." + n[len(n)-1]
			}
		}

		ext = strings.ToLower(ext)

		r := UploadResult{
			Path:     zstring.Md5Byte(b) + ext,
			Name:     v.Filename,
			Ext:      ext,
			Size:     v.Size,
			Storage:  "local",
			MimeType: mt,
			Body:     b,
		}

		if o.CustomFilter != nil {
			if err = o.CustomFilter(&r); err != nil {
				return nil, zerror.InvalidInput.Text(err.Error())
			}
		}

		uploads = append(uploads, r)
	}

	for i := range uploads {
		if o.CustomProcessing != nil {
			if err = o.CustomProcessing(&uploads[i], uploadDir); err != nil {
				return nil, zerror.InvalidInput.Text(err.Error())
			}
		} else {
			err = zfile.WriteFile(uploadDir+uploads[i].Path, uploads[i].Body)
			if err != nil {
				return nil, zerror.With(err, "文件保存失败")
			}
			uploads[i].Path = zfile.SafePath(uploadDir+uploads[i].Path, zfile.ProgramPath())
		}
	}

	return uploads, nil
}
