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
	"strings"
	"sync"

	"github.com/qor/media"
	mediaoss "github.com/qor/media/oss"
	"github.com/qor/qor/utils"
)

// Package microsite's packages struct
type Package struct {
	mediaoss.OSS
}

func (site QorMicroSite) GetPreviewURL() string {
	if site.Package.Url == "" {
		return ""
	}
	_url := strings.Replace(path.Dir(site.Package.URL()), ZIP_PACKAGE_DIR, FILE_LIST_DIR, 1)
	endPoint := mediaoss.Storage.GetEndpoint()
	endPoint = removeHttpPrefix(endPoint)

	return "//" + path.Join(endPoint, FILE_LIST_DIR, strings.Split(_url, FILE_LIST_DIR)[1], "index.html")
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
default path of package: S3Bucket/microsite/zips/id/version/
default path of files: 	 S3Bucket/microsite/id/version/
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

type fileReader struct {
	path   string
	reader *bytes.Reader
}

func UnzipPkgAndUpload(pkgURL, dest string) (files string, err error) {
	fileName, err := getFileLocalName(pkgURL)
	if err != nil {
		return files, err
	}

	if filepath.Ext(fileName) != "" {
		defer os.Remove(fileName)
	}
	reader, err := zip.OpenReader(fileName)
	if err != nil {
		return files, err
	}
	defer reader.Close()

	chFile := make(chan fileReader, CountOfThreadUpload+2)
	chErrs := make(chan error)
	var group sync.WaitGroup
	for i := 0; i < CountOfThreadUpload; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for cf := range chFile {
				if _, err0 := mediaoss.Storage.Put(cf.path, cf.reader); err0 != nil {
					chErrs <- err0
					return
				}
			}
		}()
	}

	arr := []string{}
	dest = strings.Replace(dest, ZIP_PACKAGE_DIR, FILE_LIST_DIR, 1)

Loop:
	for _, f := range reader.File {
		select {
		case err = <-chErrs:
			break Loop
		default:
			if !strings.HasPrefix(f.Name, "__MACOSX") && !strings.HasSuffix(f.Name, "DS_Store") && !f.FileInfo().IsDir() {
				var (
					rc      io.ReadCloser
					content []byte
					pth     string
				)
				rc, err = f.Open()
				if err != nil {
					break Loop
				}

				arr = append(arr, f.Name)

				content, err = ioutil.ReadAll(rc)
				if err != nil {
					rc.Close()
					break Loop
				}
				rc.Close()

				// Fix Zip Slip Vulnerability https://snyk.io/research/zip-slip-vulnerability#go
				pth, err = utils.SafeJoin(dest, f.Name)
				if err != nil {
					break Loop
				}

				chFile <- fileReader{path: pth, reader: bytes.NewReader(content)}
			}
		}
	}
	close(chFile)
	group.Wait()
	if err != nil {
		return
	}
	return strings.Join(arr, ","), nil
}

//create tempFile at locale and return its name
func getFileLocalName(path string) (fileName string, err error) {
	readCloser, err := mediaoss.Storage.GetStream(path)
	if err != nil {
		return
	}
	defer readCloser.Close()

	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("s3*%s", ext)
	if _, _err := os.Stat(TempDir); TempDir != "" && os.IsNotExist(_err) {
		if err = os.MkdirAll(TempDir, os.ModePerm); err != nil {
			return
		}
	}
	var file *os.File
	if file, err = ioutil.TempFile(TempDir, pattern); err != nil {
		return
	}
	defer file.Close()

	if _, err = io.Copy(file, readCloser); err != nil {
		return
	}

	if _, err = file.Seek(0, 0); err != nil {
		return
	}

	return file.Name(), err
}

func removeHttpPrefix(endPoint string) string {
	for _, prefix := range []string{"https://", "http://", "//"} {
		endPoint = strings.TrimPrefix(endPoint, prefix)
	}
	return endPoint
}
