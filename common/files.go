package common

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zstring"
)

type UploadOption struct {
	Key      string
	Dir      string
	MimeType []string
	MaxSize  int64
}

type UploadResult struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	MimeType string `json:"mime_type"`
	Storage  string `json:"storage"`
	Size     int64  `json:"size"`
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
	uploads := make(map[string]*multipart.FileHeader, len(files))
	buf := bytes.NewBuffer(nil)

	for _, v := range files {
		f, err := v.Open()
		if err != nil {
			return nil, invalidInput(zerror.With(err, "文件读取失败"))
		}

		if _, err := io.Copy(buf, f); err != nil {
			if err != nil {
				return nil, invalidInput(zerror.With(err, "文件读取失败"))
			}
		}

		_ = f.Close()

		b := buf.Bytes()

		if len(b) > int(o.MaxSize) {
			return nil, invalidInput(errors.New("文件大小超过限制"))
		}

		mt := zfile.GetMimeType(v.Filename, b)
		n := strings.Split(mt, "/")
		if len(n) < 2 {
			return nil, invalidInput(errors.New("文件类型错误"))
		}
		v.Header.Set("MimeType", mt)
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
				ext = "." + n[len(n)]
			}
		}

		id := zstring.Md5Byte(b) + ext
		uploads[id] = v

		buf.Reset()
	}

	res := make([]UploadResult, 0, len(uploads))
	for n, f := range uploads {
		err = c.SaveUploadedFile(f, uploadDir+n)
		if err != nil {
			return nil, zerror.With(err, "文件保存失败")
		}
		res = append(res, UploadResult{
			Path:     "/" + zfile.SafePath(uploadDir+n),
			Name:     f.Filename,
			Size:     f.Size,
			Storage:  "local",
			MimeType: strings.SplitN(f.Header.Get("MimeType"), ";", 2)[0],
		})
	}

	return res, nil

}
