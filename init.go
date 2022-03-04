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
)

const (
	Action_preview   = "preview"
	Action_publish   = "publish"
	Action_republish = "republish"
	Action_unpublish = "unpublish"

	Status_draft       = "Draft"
	Status_published   = "Published"
	Status_unpublished = "Unpublished"
)

var (
	ZIP_PACKAGE_DIR = "microsite/zips/"
	FILE_LIST_DIR   = "microsite/"
	
	//default value os.TempDir()
	TempDir             string
	CountOfThreadUpload int = 5

	prefixCollection      = []string{"/"}
	changeStatusActionMap = map[string]string{
		Action_unpublish: "Unpublished",
		Action_publish:   "Published",
		Action_republish: "Unpublished",
	}

	mdb *gorm.DB
)

func Init(s3 oss.StorageInterface, adm *admin.Admin, siteStruct QorMicroSiteInterface, admConfig *admin.Config) *admin.Resource {
	if admConfig == nil {
		admConfig = &admin.Config{Name: "MicroSites"}
	}

	mediaoss.Storage = s3

	mdb = adm.DB
	mdb.AutoMigrate(siteStruct)
	res := adm.AddResource(siteStruct, admConfig)
	ConfigureQorResource(res)
	publish2.RegisterCallbacks(mdb)
	media.RegisterCallbacks(mdb)

	return res
}

//SetPrefixCollection set collection of prefix for microsite, then select one for microsite
func SetPrefixCollection(paths []string) {
	prefixCollection = paths
}

func ConfigureQorResource(res *admin.Resource) {
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

func savehandler(oldSavehandler func(interface{}, *qor.Context) error) func(interface{}, *qor.Context) error {
	return func(resource interface{}, context *qor.Context) error {
		if site, ok := resource.(QorMicroSiteInterface); ok {
			err := oldSavehandler(resource, context)
			if err != nil {
				return err
			}

			pkg := site.GetMicroSitePackage()
			if pkg.Url == "" {
				return context.DB.Model(site).UpdateColumn("file_list", "[]").Error
			}
			if pkg.FileHeader != nil ||
				context.Request.URL.Query().Get("primary_key[qor_micro_sites_version_name]") != site.GetVersionName() { //avoid copying when normal updating
				files, err := UnzipPkgAndUpload(pkg.Url, path.Join(FILE_LIST_DIR, fmt.Sprint(site.GetId()), site.GetVersionName()))
				if err != nil {
					return err
				}

				return context.DB.Model(site).UpdateColumn("file_list", files).Error
			}
		}
		return nil
	}
}
