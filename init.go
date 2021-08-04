package microsite

import (
	"fmt"
	"html/template"
	"path"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/media"
	mediaoss "github.com/qor/media/oss"
	"github.com/qor/oss"
	"github.com/qor/publish2"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
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
	TempDir             string
	CountOfThreadUpload int = 5

	prefixCollection      = []string{"/"}
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

	return res
}

//SetPrefixCollection set collection of prefix for microsite, then select one for microsite
func SetPrefixCollection(paths []string) {
	prefixCollection = paths
}

func (site *QorMicroSite) ConfigureQorResourceBeforeInitialize(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		res.Meta(&admin.Meta{
			Name: "PrefixPath",
			Config: &admin.SelectOneConfig{
				Collection: prefixCollection,
				AllowBlank: false,
			}})
		res.Meta(&admin.Meta{
			Name: "FileList",
			Type: "readonly",
			Valuer: func(value interface{}, ctx *qor.Context) interface{} {
				site := value.(QorMicroSiteInterface)
				if site.GetStatus() == Status_unpublished {
					return ""
				}
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

				_url := strings.TrimSuffix(site.GetPreviewURL(), "/index.html")
				// List all html files first
				for _, v := range htmlFiles {
					result += fmt.Sprintf(`<br><a href="%v" target="_blank"> %v </a>`, _url+"/"+strings.TrimLeft(v, "/"), v)
				}
				// Add view all button
				result += `<br><p style='margin:10px 0'><span>Assets</span><p>`
				result += `<div>`
				for _, v := range otherFiles {
					result += fmt.Sprintf(`<a href="%v" target="_blank">%v</a><br>`, _url+"/"+strings.TrimLeft(v, "/"), v)
				}
				result += `</div>`

				return template.HTML(result)
			},
		})

		res.IndexAttrs("ID", "Name", "PrefixPath", "URL", "Status")
		res.EditAttrs("Name", "PrefixPath", "URL", "FileList", "Package")
		res.NewAttrs(res.NewAttrs(), "-Status", "-FileList")
		res.Scope(&admin.Scope{
			Name:    "",
			Default: true,
			Handler: func(db *gorm.DB, ctx *qor.Context) *gorm.DB {
				return db.Order("id DESC")
			},
		})

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
		oldSavehandler := res.SaveHandler
		res.SaveHandler = savehandler(oldSavehandler)
	}
}

func savehandler(oldSavehandler func(interface{}, *qor.Context) error) func(resource interface{}, context *qor.Context) error {
	return func(resource interface{}, context *qor.Context) error {
		if site, ok := resource.(QorMicroSiteInterface); ok {
			err := oldSavehandler(resource, context)
			if err != nil {
				return err
			}

			pkg := site.GetMicroSitePackage()
			if pkg.Url == "" {
				site.SetFileList("[]")
				return context.DB.Save(site).Error
			}

			if pkg.FileHeader != nil ||
				context.Request.URL.Query().Get("primary_key[qor_micro_sites_version_name]") != site.GetVersionName() { //avoid copying when normal updating
				files, err := UnzipPkgAndUpload(pkg.Url, path.Join(FILE_LIST_DIR, fmt.Sprint(site.GetId()), site.GetVersionName()))
				if err != nil {
					return err
				}
				site.SetFileList(files)
				return context.DB.Save(site).Error
			}
		}
		return nil
	}
}
