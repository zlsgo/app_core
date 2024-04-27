package common

import (
	"bytes"
	"errors"
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
	CustomProcessing func(r *UploadResult) error
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

	if subDirName != "" {
		subDirName = strings.Replace(subDirName, "::", "__", -1)
		subDirName = strings.Trim(subDirName, "/")
		subDirName = "/" + subDirName
	}
	uploadDir := ""
	if o.Dir != "" {
		uploadDir = zfile.RealPathMkdir(Define.UploadLocalDir+o.Dir, true)
		if !strings.HasPrefix(o.Dir, "/") {
			uploadDir = zfile.RealPathMkdir(Define.UploadLocalDir+subDirName+"/"+o.Dir, true)
		}
	} else {
		uploadDir = zfile.RealPathMkdir(Define.UploadLocalDir+subDirName, true)
	}

	invalidInput := zerror.WrapTag(zerror.InvalidInput)
	uploads := make([]UploadResult, 0, len(files))
	buf := bytes.NewBuffer(nil)

	for _, v := range files {
		f, err := v.Open()
		if err != nil {
			return nil, invalidInput(zerror.With(err, "文件读取失败"))
		}

		if _, err := io.Copy(buf, f); err != nil {
			return nil, invalidInput(zerror.With(err, "文件读取失败"))
		}

		_ = f.Close()

		b := buf.Bytes()

		if len(b) > int(o.MaxSize) {
			return nil, invalidInput(errors.New("文件大小超过限制"))
		}

		mt := zfile.GetMimeType(v.Filename, b)
		n := strings.Split(mt, "/")
		if len(n) < 2 {
			return nil, invalidInput(errors.New("文件类型无法识别"))
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
				return nil, invalidInput(errors.New("不支持的文件类型"))
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
				return nil, invalidInput(err)
			}
		}

		uploads = append(uploads, r)
		buf.Reset()
	}

	for i := range uploads {
		if o.CustomProcessing != nil {
			if err = o.CustomProcessing(&uploads[i]); err != nil {
				return nil, invalidInput(err)
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
