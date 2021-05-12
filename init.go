package microsite

import (
	"fmt"
	"html/template"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/media"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/roles"
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
var admDB *gorm.DB
var TempDir string = "public/system/qor_jobs"

func init() {
	media.RegisterMediaHandler("unzip_package_handler", unzipPackageHandler{})
}

func (site *QorMicroSite) ConfigureQorResourceBeforeInitialize(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		if admDB == nil {
			admDB = res.GetAdmin().DB
		}

		res.Meta(&admin.Meta{Name: "Name", Label: "Site Name", Permission: roles.Deny(roles.Update, roles.Anyone)})
		res.Meta(&admin.Meta{Name: "URL", Label: "Microsite URL", Permission: roles.Deny(roles.Update, roles.Anyone)})
		res.Meta(&admin.Meta{
			Name: "FileList",
			Type: "readonly",
			Valuer: func(value interface{}, ctx *qor.Context) interface{} {
				this := value.(QorMicroSiteInterface)
				var result string
				for _, v := range this.GetFileList() {
					result += fmt.Sprintf(`<a href="%v" target="_blank"> %v </a><br>`, this.GetPreviewURL()+"/"+v, v)
				}
				return template.HTML(result)
			},
		})

		res.IndexAttrs("Name", "URL", "Status")
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
}
