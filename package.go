package microsite

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/qor/media"
	mediaoss "github.com/qor/media/oss"
	"github.com/qor/oss"
	"github.com/qor/qor/utils"
)

// Package microsite's packages struct
type Package struct {
	mediaoss.OSS
}

// ListObjects list all objects under current path
func (pkg Package) ListObjects() ([]*oss.Object, error) {
	return mediaoss.Storage.List(filepath.Dir(pkg.URL()))
}

func (site QorMicroSite) GetPreviewURL() string {
	if site.Package.Url == "" {
		return ""
	}
	_url := strings.Replace(path.Dir(site.Package.URL()), ZIP_PACKAGE_DIR, FILE_LIST_DIR, 1)
	endPoint := mediaoss.Storage.GetEndpoint()
	endPoint = removeHttpPrefix(endPoint)

	return "//" + path.Join(endPoint, FILE_LIST_DIR, strings.Split(_url, FILE_LIST_DIR)[1])
}

// unzipPackageHandler unzip microsite package
type unzipPackageHandler struct {
}

func (packageHandler unzipPackageHandler) CouldHandle(media media.Media) bool {
	if _, ok := media.(*Package); ok {
		return true
	}
	return false
}

/*
the path of package: /privates3/microsite/zips/id/version/
the path of files: 	 /privates3/microsite/id/version/
*/
func (packageHandler unzipPackageHandler) Handle(media media.Media, file media.FileInterface, option *media.Option) (err error) {
	if pkg, ok := media.(*Package); ok && file != nil {
		fileURL := media.URL()
		fileURL = strings.TrimPrefix(strings.TrimLeft(fileURL, "/"), removeHttpPrefix(mediaoss.Storage.GetEndpoint()))
		if err = media.Store(fileURL, option, file); err == nil {
			if pkg.Options == nil {
				pkg.Options = map[string]string{}
			}
			pkg.Options["file_list"], err = UnzipPkgAndUpload(fileURL, filepath.Dir(fileURL))
			return err
		}
	}
	return err
}

func UnzipPkgAndUpload(pkgURL, dest string) (files string, err error) {
	baseName := strings.TrimSuffix(filepath.Base(pkgURL), filepath.Ext(pkgURL))
	file, err := getFile(pkgURL)
	if err != nil {
		return files, err
	}

	if filepath.Ext(file.Name()) != "" {
		defer os.Remove(file.Name())
	}
	reader, err := zip.OpenReader(file.Name())
	if err != nil {
		return files, err
	}
	defer reader.Close()

	filePrefix := baseName
	{
		folders := []string{}
		for _, f := range reader.File {
			if !strings.HasPrefix(f.Name, "__MACOSX") && f.FileInfo().IsDir() {
				folders = append(folders, f.Name)
			}
		}
		sort.Strings(folders)
		if len(folders) > 0 {
			matched := true
			newPrefix := folders[0]
			for _, folder := range folders {
				if !strings.HasPrefix(folder, newPrefix) {
					matched = false
					break
				}
			}

			if matched {
				filePrefix = newPrefix
			}
		}
	}
	arr := []string{}
	dest = strings.Replace(dest, ZIP_PACKAGE_DIR, FILE_LIST_DIR, 1)
	for _, f := range reader.File {
		if !strings.HasPrefix(f.Name, "__MACOSX") && !strings.HasSuffix(f.Name, "DS_Store") && !f.FileInfo().IsDir() {
			rc, err := f.Open()
			if err != nil {
				return files, err
			}
			defer rc.Close()

			fixedFileName := strings.TrimPrefix(f.Name, filePrefix)
			arr = append(arr, fixedFileName)
			content, err := ioutil.ReadAll(rc)
			if err != nil {
				return files, err
			}

			// Fix Zip Slip Vulnerability https://snyk.io/research/zip-slip-vulnerability#go
			if pth, err := utils.SafeJoin(dest, fixedFileName); err == nil {
				if _, err := mediaoss.Storage.Put(pth, bytes.NewReader(content)); err != nil {
					return files, err
				}
			} else {
				return files, err
			}
		}
	}

	return strings.Join(arr, ","), nil
}

//create tempFile at locale
func getFile(path string) (file *os.File, err error) {
	readCloser, err := mediaoss.Storage.GetStream(path)
	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("s3*%s", ext)
	if _, _err := os.Stat(TempDir); TempDir != "" && os.IsNotExist(_err) {
		err = os.MkdirAll(TempDir, os.ModePerm)
	}
	if err == nil {
		if file, err = ioutil.TempFile(TempDir, pattern); err == nil {
			defer readCloser.Close()
			_, err = io.Copy(file, readCloser)
			file.Seek(0, 0)
		}
	}

	return file, err
}

func removeHttpPrefix(endPoint string) string {
	for _, prefix := range []string{"https://", "http://", "//"} {
		endPoint = strings.TrimPrefix(endPoint, prefix)
	}
	return endPoint
}
