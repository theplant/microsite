package microsite_test

import (
	"testing"

	"github.com/qor/admin"
	"github.com/qor/qor/test/utils"
	"github.com/theplant/microsite"
)

type TestMicroSite struct {
	microsite.QorMicroSite
}

func TestInit(t *testing.T) {
	adm := SetupAdmin()
	s3 := InitTestS3()
	microsite.Init(s3, adm, &TestMicroSite{}, &admin.Config{Name: "TestMicroSites"})

	res := adm.GetResource("TestMicroSites")
	if res == nil {
		t.Error("microsite is not registered")
	}
}

func SetupAdmin() *admin.Admin {
	db := utils.GetTestDB()
	adm := admin.New(&admin.AdminConfig{SiteName: "microsite test", DB: db})

	return adm
}
