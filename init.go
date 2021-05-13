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

	Action_preview   = "preview"
	Action_publish   = "publish"
	Action_republish = "republish"
	Action_unpublish = "unpublish"

	Status_draft       = "Draft"
	Status_published   = "Published"
	Status_unpublished = "Unpublished"
)

var (
	//default value os.TempDir()
	TempDir string
	admDB   *gorm.DB

	changeStatusActionMap = map[string]string{
		Action_unpublish: "Unpublished",
		Action_publish:   "Published",
		Action_republish: "Unpublished",
	}
)

func Init(adm *admin.Admin, siteStruct QorMicroSiteInterface) {
	db := adm.DB
	db.AutoMigrate(siteStruct)
	adm.AddResource(siteStruct, &admin.Config{Name: "MicroSites"})

	publish2.RegisterCallbacks(db)
	media.RegisterCallbacks(db)
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
				site := value.(QorMicroSiteInterface)
				var result string
				for _, v := range site.GetFileList() {
					result += fmt.Sprintf(`<a href="%v" target="_blank"> %v </a><br>`, site.GetPreviewURL()+"/"+v, v)
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
				site := record.(QorMicroSiteInterface)
				return site.GetPreviewURL()
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
