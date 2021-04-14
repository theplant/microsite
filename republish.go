package microsite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/oss"
	"github.com/qor/publish2"
	"github.com/theplant/appkit/db"
	"github.com/theplant/gormutils"
)

func Republish(ctx context.Context, version QorMicroSiteInterface, printActivityLog bool) (err error) {
	_db := db.MustGetGorm(ctx)
	tableName := _db.NewScope(version).TableName()

	err = gormutils.Transact(_db, func(tx *gorm.DB) (err1 error) {
		defer func() {
			if err1 != nil {
				eventType := fmt.Sprintf("%s:RepublishingError", strings.Title(tableName))
				fmt.Printf("%v, error: %v\n", eventType, err1.Error())
			}
		}()

		var liveRecord QorMicroSite
		tx.Set(publish2.VersionMode, publish2.VersionMultipleMode).Set(publish2.ScheduleMode, publish2.ModeOff).Where("id = ? AND status = ?", version.GetId(), Status_published).First(&liveRecord)

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
		version.SetVersionPriority(fmt.Sprintf("%v", time.Now().UTC().Format(time.RFC3339)) + "_" + version.GetVersionName())
		if err1 = tx.Save(version).Error; err1 != nil {
			return
		}

		if err1 = version.SitemapHandler(_db, version.GetMicroSiteURL(), Action_republish); err1 != nil {
			return
		}

		_, err1 = UnzipPkgAndUpload(version.GetMicroSitePackage().Url, version.GetMicroSiteURL())
		return
	})

	return
}
