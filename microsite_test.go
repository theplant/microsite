package microsite_test

import (
	"testing"

	"github.com/theplant/microsite"
)

func TestGetFilesPreviewURL(t *testing.T) {
	m := NewMicrosite(microsite.Status_draft)
	m.ID = 5
	m.Package.FileName = "sss.zip"
	m.Package.Url = "/microsite/zips/5/20210419152358/sss.zip"
	m.Package.Options = map[string]string{"file_list": "index.html,js/main.js,check/check.go,check/subcheck/sub.css"}

	arr := m.GetFileList()
	prefix := "/microsite/5/20210419152358/"
	for _, s := range []string{
		"",
		"microsite/zips/5/20210419152358/sss.zip",
		"/microsite/zips/5/20210419152358/sss.zip",
		"//northeast-1.amazonaws.com/microsite/zips/5/20210419152358/sss.zip"} {
		m.Package.Url = s
		for k, v := range m.GetFilesPreviewURL() {
			w := prefix + arr[k]
			if v != w {
				t.Errorf("want %v, got %v", w, v)
			} else {
				t.Log(v)
			}
		}
	}
}
