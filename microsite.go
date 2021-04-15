package microsite

import (
	"context"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/publish2"
)

// QorMicroSiteInterface defined QorMicroSite itself's interface
type QorMicroSiteInterface interface {
	GetId() uint
	GetMicroSiteURL() string
	GetMicroSitePackage() *Package
	GetFileList() string
	GetPreviewURL() string
	GetVersionName() string
	SetVersionPriority(string)
	GetStatus() string
	SetStatus(string)
	TableName() string
	GetCreatedAt() time.Time
	PublishCallBack(db *gorm.DB, sitePath string) error
	UnPublishCallBack(db *gorm.DB, sitePath string) error
}

// QorMicroSite default qor microsite setting struct
type QorMicroSite struct {
	gorm.Model

	publish2.Version
	publish2.Schedule

	Name    string
	URL     string
	Status  string
	Package Package `gorm:"size:65536" media_library:"url:/microsite/zips/{{primary_key}}/{{short_hash}}/{{filename}}"`
}

// GetMicroSiteID will return a site's ID
func (site QorMicroSite) GetId() uint {
	return site.ID
}

func (site QorMicroSite) GetFileList() string {
	return site.Package.Options["file_list"]
}

// GetMicroSiteURL will return a site's URL
func (site QorMicroSite) GetMicroSiteURL() string {
	return site.URL
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
	return
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

func (this *QorMicroSite) BeforeDelete(db *gorm.DB) (err error) {
	if this.Status == Status_published {
		err = Unpublish(context.TODO(), this, false)
		return
	}
	return
}
