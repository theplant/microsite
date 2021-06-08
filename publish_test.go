package microsite_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/qor/oss"
)

type MockQORS3Client struct {
	oss.StorageInterface
	Objects map[string]string
}

var MockS3 *MockQORS3Client

func HasFileInS3(filePath string) bool {
	return len(MockS3.Objects[filePath]) != 0
}

func InitTestS3() *MockQORS3Client {
	mockS3 := MockQORS3Client{}

	return &mockS3
}

func (m *MockQORS3Client) Get(path string) (*os.File, error) {
	// If we can find the "file" in given path, return nil error to
	// indicate the Get call is succeed
	if _, ok := m.Objects[path]; ok {
		return nil, nil
	}

	return nil, errors.New("no file")
}

func (m *MockQORS3Client) Put(path string, r io.Reader) (*oss.Object, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}

	if m.Objects == nil {
		m.Objects = make(map[string]string)
	}

	m.Objects[path] = string(b)

	return &oss.Object{Path: path}, nil
}

func (m *MockQORS3Client) Delete(path string) error {
	if m.Objects == nil {
		m.Objects = make(map[string]string)
	}
	if _, ok := m.Objects[path]; ok {
		delete(m.Objects, path)
	}

	return nil
}
