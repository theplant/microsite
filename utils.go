package microsite

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/qor/publish2"
)

func AutoPublishMicrosite(db *gorm.DB, readyForPublishStatus string) error {
	ctx := context.TODO()
	sites := []QorMicroSite{}

	if err := db.Set(publish2.VersionMode, publish2.VersionMultipleMode).Set(publish2.ScheduleMode, publish2.ComingOnlineMode).
		Where("status = ?", readyForPublishStatus).Find(&sites).Error; err != nil {
		return err
	}

	for _, site := range sites {
		if err := Publish(ctx, &site, true); err != nil {
			return err
		}
	}
	return nil
}

func AutoUnpublishMicrosite(db *gorm.DB, unPublishStatus string) error {
	ctx := context.TODO()
	sites := []QorMicroSite{}

	if err := db.Set(publish2.VersionMode, publish2.VersionMultipleMode).Set(publish2.ScheduleMode, publish2.GoingOfflineMode).
		Where("status = ?", unPublishStatus).Find(&sites).Error; err != nil {
		return err
	}

	for _, site := range sites {
		if err := Unpublish(ctx, &site, true); err != nil {
			return err
		}
	}
	return nil
}
