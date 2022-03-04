package microsite

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/publish2"
)

type GetSiteFunc func(db *gorm.DB, targetStatus string) ([]QorMicroSiteInterface, error)

func ToPublishMicrosites(db *gorm.DB, readyForPublishStatus string, fn GetSiteFunc) error {
	sites, err := fn(db, readyForPublishStatus)
	if err != nil {
		return err
	}

	for _, site := range sites {
		if err := Publish(db, site, nil); err != nil {
			return err
		}
	}
	return nil
}

func ToUnpublishMicrosites(db *gorm.DB, unPublishStatus string, fn GetSiteFunc) error {
	sites, err := fn(db, unPublishStatus)
	if err != nil {
		return err
	}

	for _, site := range sites {
		if err := Unpublish(db, site, nil); err != nil {
			return err
		}
	}
	return nil
}

func GetSites(db *gorm.DB, targetStatus string) (arr []QorMicroSiteInterface, err error) {
	sites := []QorMicroSite{}

	if err := db.Set(publish2.VersionMode, publish2.VersionMultipleMode).Where("status = ?", targetStatus).Find(&sites).Error; err != nil {
		return arr, err
	}

	for _, v := range sites {
		arr = append(arr, &v)
	}
	return
}
