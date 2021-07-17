package microsite

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/publish2"
)

// QorMicroSiteInterface defined QorMicroSite itself's interface
type QorMicroSiteInterface interface {
	GetId() uint
	GetMicroSiteURL() string
	GetPrefixPath() string
	GetMicroSitePackage() *Package
	GetFileList() []string
	GetFilesPathWithSiteURL() []string
	GetFilesPreviewURL() []string
	GetPreviewURL() string
	GetVersionName() string
	SetVersionPriority(string)
	GetStatus() string
	TableName() string
	GetCreatedAt() time.Time
	PublishCallBack(db *gorm.DB, sitePath string) error
	UnPublishCallBack(db *gorm.DB, sitePath string) error

	SetStatus(string)
	SetScheduledEndAt(*time.Time)
}

// QorMicroSite default qor microsite setting struct
type QorMicroSite struct {
	gorm.Model

	publish2.Version
	publish2.Schedule

	Name       string
	PrefixPath string
	URL        string
	Status     string
	Package    Package `gorm:"size:65536" media_library:"url:/microsite/zips/{{primary_key}}/{{short_hash}}/{{filename}}"`
}

// GetMicroSiteID will return a site's ID
func (site QorMicroSite) GetId() uint {
	return site.ID
}

func (site QorMicroSite) GetFileList() (arr []string) {
	return strings.Split(site.Package.Options["file_list"], ",")
}

func (site QorMicroSite) GetFilesPathWithSiteURL() (arr []string) {
	for _, v := range site.GetFileList() {
		arr = append(arr, path.Join(site.GetMicroSiteURL(), v))
	}
	return
}

func (site QorMicroSite) GetFilesPreviewURL() (arr []string) {
	if site.Package.URL() != "" {
		_url := "/" + FILE_LIST_DIR + strings.Split(path.Dir(site.Package.URL()), ZIP_PACKAGE_DIR)[1]
		for _, v := range site.GetFileList() {
			arr = append(arr, path.Join(_url, v))
		}
	}
	return
}

// GetMicroSiteURL will return a site's URL
func (site QorMicroSite) GetMicroSiteURL() string {
	return path.Join(site.PrefixPath, site.URL)
}

func (site QorMicroSite) GetPrefixPath() string {
	return site.PrefixPath
}

// GetMicroSitePackage get microsite package
func (site QorMicroSite) GetMicroSitePackage() *Package {
	return &site.Package
}

func (site *QorMicroSite) TableName() string {
	return "qor_micro_sites"
}

func (site QorMicroSite) GetVersionName() string {
	return site.VersionName
}

func (site *QorMicroSite) SetVersionPriority(versionPriority string) {
	site.VersionPriority = versionPriority
}

func (site QorMicroSite) GetCreatedAt() time.Time {
	return site.CreatedAt
}

func (site QorMicroSite) GetStatus() string {
	return site.Status
}

func (site *QorMicroSite) SetStatus(status string) {
	site.Status = status
}

func (site *QorMicroSite) SetScheduledEndAt(t *time.Time) {
	site.ScheduledEndAt = t
}

func (site QorMicroSite) PublishCallBack(db *gorm.DB, sitePath string) error {
	return nil
}

func (site QorMicroSite) UnPublishCallBack(db *gorm.DB, sitePath string) error {
	return nil
}

func (site *QorMicroSite) BeforeCreate(db *gorm.DB) (err error) {
	site.Status = Status_draft
	site.CreatedAt = gorm.NowFunc()
	site.VersionPriority = fmt.Sprintf("%v", site.CreatedAt.UTC().Format(time.RFC3339))
	return nil
}

func (site *QorMicroSite) BeforeUpdate(db *gorm.DB) (err error) {
	if site.Status == Status_published {
		site.VersionPriority = fmt.Sprintf("%v", gorm.NowFunc().UTC().Format(time.RFC3339))
	} else {
		site.VersionPriority = fmt.Sprintf("%v", site.CreatedAt.UTC().Format(time.RFC3339))
	}
	return nil
}

func (site *QorMicroSite) BeforeDelete(db *gorm.DB) (err error) {
	if site.Status == Status_published {
		err = Unpublish(db, site, false)
		return
	}

	return
}
