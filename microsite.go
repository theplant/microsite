package microsite

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/oss"
	"github.com/qor/publish2"
)

// QorMicroSiteInterface defined QorMicroSite itself's interface
type QorMicroSiteInterface interface {
	GetId() uint
	GetMicroSiteURL() string
	GetPrefixPath() string
	GetMicroSitePackage() *oss.OSS
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

	SetFileList(string)
	SetStatus(string)
	SetScheduledStartAt(*time.Time)
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
	FileList   string
	Package    oss.OSS `gorm:"size:65536" media_library:"url:/microsite/zips/{{primary_key}}/{{short_hash}}/{{filename}}"`
}

// GetMicroSiteID will return a site's ID
func (site QorMicroSite) GetId() uint {
	return site.ID
}

func (site QorMicroSite) GetFileList() (arr []string) {
	json.Unmarshal([]byte(site.FileList), &arr)
	return
}

func (site QorMicroSite) GetFilesPathWithSiteURL() (arr []string) {
	for _, v := range site.GetFileList() {
		arr = append(arr, path.Join(site.GetMicroSiteURL(), v))
	}
	return
}

func (site QorMicroSite) GetFilesPreviewURL() (arr []string) {
	if site.Package.URL() != "" {
		_url := path.Join("/"+FILE_LIST_DIR, fmt.Sprint(site.ID), site.VersionName)
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

// // GetMicroSitePackage get microsite package
func (site QorMicroSite) GetMicroSitePackage() *oss.OSS {
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

func (site *QorMicroSite) SetScheduledStartAt(t *time.Time) {
	site.ScheduledStartAt = t
}

func (site *QorMicroSite) SetScheduledEndAt(t *time.Time) {
	site.ScheduledEndAt = t
}

func (site *QorMicroSite) SetFileList(s string) {
	site.FileList = s
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
	} else if site.Status != Status_unpublished { //draft,approved,review
		//clear preview files
		if s3, ok := oss.Storage.(DeleteObjecter); ok {
			err = s3.DeleteObjects(site.GetFilesPreviewURL())
		} else {
			for _, o := range site.GetFilesPreviewURL() {
				oss.Storage.Delete(o)
			}
		}
	}

	return
}

func (site QorMicroSite) GetPreviewURL() string {
	if site.Package.Url == "" {
		return ""
	}
	endPoint := oss.Storage.GetEndpoint()
	endPoint = removeHttpPrefix(endPoint)

	return "//" + path.Join(endPoint, FILE_LIST_DIR, fmt.Sprint(site.ID), site.VersionName, "index.html")
}
