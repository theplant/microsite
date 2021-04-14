package microsite

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"github.com/qor/admin"
	"github.com/qor/media"
	"github.com/qor/oss"
	"github.com/qor/publish2"
	"github.com/qor/qor"
)

var (
	publicS3  oss.StorageInterface
	privateS3 oss.StorageInterface
	DB        *gorm.DB
)

const (
	ZIP_PACKAGE_DIR = "microsite/zips/"
	FILE_LIST_DIR   = "microsite/"

	Action_preview        = "preview"
	Action_request_review = "request review"
	Action_approve        = "approve"
	Action_return         = "return"
	Action_publish        = "publish"
	Action_republish      = "republish"
	Action_unpublish      = "unpublish"

	Status_draft       = "Draft"
	Status_review      = "Review"
	Status_approved    = "Approved"
	Status_returned    = "Returned"
	Status_published   = "Published"
	Status_unpublished = "Unpublished"
)

var changeStatusActionMap = map[string]string{
	Action_request_review: "Review",
	Action_approve:        "Approved",
	Action_return:         "Returned",
	Action_unpublish:      "Unpublished",
	Action_publish:        "Published",
	Action_republish:      "Unpublished",
}

//New initialize a microsite
func New(adm *admin.Admin) *admin.Resource {
	inflection.AddUncountable("micro_sites")
	if err := os.MkdirAll("public/system/qor_jobs", os.ModePerm); err != nil && !os.IsExist(err) && !os.IsPermission(err) {
		panic(err)
	}
	DB := adm.DB
	DB.AutoMigrate(&QorMicroSite{})
	media.RegisterMediaHandler("unzip_package_handler", unzipPackageHandler{})
	return AddAdminResource(adm, "MicroSites")
}

func AddAdminResource(adm *admin.Admin, menuName string) *admin.Resource {
	res := adm.AddResource(&QorMicroSite{}, &admin.Config{Name: menuName})
	if menuName == "MicroSites" {
		res.UseTheme("versions")
	}

	res.Meta(&admin.Meta{Name: "Name", Label: "Site Name"})
	res.Meta(&admin.Meta{Name: "URL", Label: "Microsite URL"})
	res.Meta(&admin.Meta{
		Name: "FileList",
		Type: "readonly",
		Valuer: func(value interface{}, ctx *qor.Context) interface{} {
			this := value.(QorMicroSiteInterface)
			var result string
			for _, v := range strings.Split(this.GetFileList(), ",") {
				result += fmt.Sprintf(`<a href="%v" target="_blank"> %v </a><br>`, this.GetPreviewURL()+"/"+v, v)
			}
			return template.HTML(result)
		},
	})

	res.IndexAttrs("Name", "VersionName", "URL", "PublishedAt", "Status")
	res.EditAttrs("Name", "URL", "FileList", "Package")
	res.NewAttrs(res.NewAttrs(), "-Status", "-FileList")

	res.Action(&admin.Action{
		Name: "Preview",
		URL: func(record interface{}, context *admin.Context) string {
			this := record.(QorMicroSiteInterface)
			return this.GetPreviewURL()
		},
		URLOpenType: "_blank",
		Modes:       []string{"show", "edit"},
	})

	res.Action(&admin.Action{
		Name: "Publish",
		Handler: func(argument *admin.ActionArgument) (err error) {
			err = ChangeStatus(argument, Action_publish)
			return
		},
		Modes: []string{"edit"},
	})

	res.Action(&admin.Action{
		Name: "Republish",
		Handler: func(argument *admin.ActionArgument) (err error) {
			err = ChangeStatus(argument, Action_republish)
			return
		},
		Modes: []string{"edit"},
	})

	res.Action(&admin.Action{
		Name: "Unpublish",
		Handler: func(argument *admin.ActionArgument) (err error) {
			err = ChangeStatus(argument, Action_unpublish)
			return
		},
		Modes: []string{"edit"},
	})
	res.OverrideIndexAttrs(func() {
		res.IndexAttrs(res.IndexAttrs(), "-PublishLiveNow", "-ScheduledStartAt", "-ScheduledEndAt")
	})
	return res
}

func AutoPublishMicrosite() error {
	ctx := context.TODO()
	sites := []QorMicroSite{}
	DB.Set(publish2.VersionMode, publish2.VersionMultipleMode).Set(publish2.ScheduleMode, publish2.ModeOff).Where("status = ? AND scheduled_start_at <= ?", Status_approved, gorm.NowFunc().Add(time.Minute)).Find(&sites)
	for _, version := range sites {
		if err := Publish(ctx, &version, true); err != nil {
			return err
		}
	}
	return nil
}
