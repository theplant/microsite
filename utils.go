package microsite

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/publish2"
)

func TobePublishedMicrosites(db *gorm.DB, readyForPublishStatus string) error {
	sites, err := GetSites(db, readyForPublishStatus)
	if err != nil {
		return err
	}

	for _, site := range sites {
		if err := Publish(db, &site, true); err != nil {
			return err
		}
	}
	return nil
}

func TobeUnpublishedMicrosite(db *gorm.DB, unPublishStatus string) error {
	sites, err := GetSites(db, unPublishStatus)
	if err != nil {
		return err
	}

	for _, site := range sites {
		if err := Unpublish(db, &site, true); err != nil {
			return err
		}
	}
	return nil
}

func GetSites(db *gorm.DB, targetStatus string) ([]QorMicroSite, error) {
	sites := []QorMicroSite{}

	if err := db.Set(publish2.VersionMode, publish2.VersionMultipleMode).Where("status = ?", targetStatus).Find(&sites).Error; err != nil {
		return []QorMicroSite{}, err
	}

	return sites, nil
}
