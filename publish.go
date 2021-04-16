package microsite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/oss"
	"github.com/qor/publish2"
	"github.com/theplant/gormutils"
)

func Publish(ctx context.Context, version QorMicroSiteInterface, printActivityLog bool) (err error) {
	_db := ctx.Value("DB").(*gorm.DB)
	tableName := _db.NewScope(version).TableName()

	err = gormutils.Transact(_db, func(tx *gorm.DB) (err1 error) {
		defer func() {
			if err1 != nil {
				eventType := fmt.Sprintf("%s:PublishingError", strings.Title(tableName))
				fmt.Printf("%v, error: %v\n", eventType, err1.Error())
			}
		}()

		var liveRecord QorMicroSite
		scope := tx.Set(publish2.VersionMode, publish2.VersionMultipleMode).Set(publish2.ScheduleMode, publish2.ModeOff).Where("id = ? AND status = ?", version.GetId(), Status_published)
		if version.GetVersionName() != "" {
			scope = scope.Where("version_name <> ?", version.GetVersionName())
		}
		scope.First(&liveRecord)

		if liveRecord.GetId() != 0 {
			objs, _ := oss.Storage.List(liveRecord.GetMicroSiteURL())
			for _, o := range objs {
				oss.Storage.Delete(o.Path)
			}

			liveRecord.SetStatus(Status_unpublished)
			liveRecord.SetVersionPriority(fmt.Sprintf("%v", liveRecord.GetCreatedAt().UTC().Format(time.RFC3339)))
			if err1 = tx.Save(&liveRecord).Error; err1 != nil {
				return
			}
		}

		version.SetStatus(Status_published)
		version.SetVersionPriority(fmt.Sprintf("%v", time.Now().UTC().Format(time.RFC3339)))
		if err1 = tx.Save(version).Error; err1 != nil {
			return
		}

		if _, err1 = UnzipPkgAndUpload(version.GetMicroSitePackage().Url, version.GetMicroSiteURL()); err1 != nil {
			return
		}

		return version.PublishCallBack(_db, version.GetMicroSiteURL())
	})

	return
}
