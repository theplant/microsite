package microsite

import (
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/qor/admin"
	"github.com/qor/media"
	mediaoss "github.com/qor/media/oss"
	"github.com/qor/oss"
	"github.com/qor/publish2"
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

	changeStatusActionMap = map[string]string{
		Action_unpublish: "Unpublished",
		Action_publish:   "Published",
		Action_republish: "Unpublished",
	}
)

func Init(s3 oss.StorageInterface, adm *admin.Admin, siteStruct QorMicroSiteInterface, admConfig *admin.Config) *admin.Resource {
	if admConfig == nil {
		admConfig = &admin.Config{Name: "MicroSites"}
	}

	mediaoss.Storage = s3

	db := adm.DB
	db.AutoMigrate(siteStruct)
	res := adm.AddResource(siteStruct, admConfig)

	publish2.RegisterCallbacks(db)
	media.RegisterCallbacks(db)
	media.RegisterMediaHandler("unzip_package_handler", unzipPackageHandler{})

	return res
}

func (site *QorMicroSite) ConfigureQorResourceBeforeInitialize(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		res.Meta(&admin.Meta{Name: "Name", Label: "Site Name", Permission: roles.Deny(roles.Update, roles.Anyone)})
		res.Meta(&admin.Meta{Name: "URL", Label: "Microsite URL", Permission: roles.Deny(roles.Update, roles.Anyone)})
		res.Meta(&admin.Meta{
			Name: "FileList",
			Type: "readonly",
			Valuer: func(value interface{}, ctx *qor.Context) interface{} {
				site := value.(QorMicroSiteInterface)
				var result string
				htmlFiles := []string{}
				otherFiles := []string{}

				for _, v := range site.GetFileList() {
					if filepath.Ext(v) == ".html" {
						htmlFiles = append(htmlFiles, v)
					} else {
						otherFiles = append(otherFiles, v)
					}
				}

				// List all html files first
				for _, v := range htmlFiles {
					result += fmt.Sprintf(`<br><a href="%v" target="_blank"> %v </a>`, site.GetPreviewURL()+"/"+v, v)
				}
				// Add view all button
				result += `<br><p style='margin:10px 0'><span>Assets</span><p>`
				result += `<div>`
				for _, v := range otherFiles {
					result += fmt.Sprintf(`<a href="%v" target="_blank">%v</a><br>`, site.GetPreviewURL()+"/"+v, v)
				}
				result += `</div>`

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
