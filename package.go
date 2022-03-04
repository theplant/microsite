package microsite

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	mediaoss "github.com/qor/media/oss"
	"github.com/qor/qor/utils"
)

/*
default path of package: S3Bucket/microsite/zips/id/hash/
default path of files: 	 S3Bucket/microsite/id/version/
*/
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

	data, err := json.Marshal(arr)
	return string(data), err
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
