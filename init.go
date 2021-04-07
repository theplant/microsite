package microsite

import (
	"context"
	"html/template"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"github.com/qor/admin"
	"github.com/qor/media"
	"github.com/qor/oss"
	"github.com/qor/publish2"
	"github.com/qor/qor"
	"github.com/qor/roles"
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
func New(adm *admin.Admin, pubS3 oss.StorageInterface, priS3 oss.StorageInterface) {
	publicS3 = pubS3
	privateS3 = priS3

	inflection.AddUncountable("micro_sites")
	inflection.AddUncountable("microsite_versions")

	DB := adm.DB
	DB.AutoMigrate(&QorMicroSite{})
	addAdminResource(adm, "MicroSites")
	addAdminResource(adm, "microsite_versions")
	adm.GetMenu("microsite_versions").Permission = roles.Deny(roles.CRUD, roles.Anyone)
	media.RegisterMediaHandler("unzip_package_handler", unzipPackageHandler{})
}

func addAdminResource(adm *admin.Admin, name string) {
	res := adm.AddResource(&QorMicroSite{}, &admin.Config{Name: name})
	if name != "microsite_versions" {
		res.UseTheme("general_resource")
		res.IndexAttrs("ID", "Name", "URL", "PublishedAt", "Live")
		res.Meta(&admin.Meta{
			Name: "Live",
			Valuer: func(record interface{}, ctx *qor.Context) interface{} {
				if ann, ok := record.(interface {
					GetMicroSiteID() uint
				}); ok {
					var count int
					ctx.DB.Model(&QorMicroSite{}).Set(publish2.VersionMode, publish2.VersionMultipleMode).Set(publish2.ScheduleMode, publish2.ModeOff).Where("id = ? AND status = ?", ann.GetMicroSiteID(), Status_published).Count(&count)
					if count > 0 {
						return template.HTML("<span class='qor-publish2__live'><span class='qor-symbol qor-symbol-blue' title='This page is live'></span></span>")
					}
					return ""
				}
				return ""
			},
		})
	} else {
		res.UseTheme("versions")
		res.IndexAttrs("Name", "URL", "PublishedAt", "Status")
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
				result += v + "<br>"
			}
			return template.HTML(result)
		},
	})

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
