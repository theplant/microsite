package microsite

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/oss"
	"github.com/qor/publish2"
	"github.com/theplant/gormutils"
)

func Republish(ctx context.Context, version QorMicroSiteInterface, printActivityLog bool) (err error) {
	_db := ctx.Value("DB").(*gorm.DB)
	tableName := _db.NewScope(version).TableName()

	err = gormutils.Transact(_db, func(tx *gorm.DB) (err1 error) {
		defer func() {
			if err1 != nil {
				eventType := fmt.Sprintf("%s:RepublishingError", strings.Title(tableName))
				fmt.Printf("%v, error: %v\n", eventType, err1.Error())
			}
		}()

		iRecord := reflect.New(reflect.TypeOf(version).Elem()).Interface()
		admDB.Set(publish2.VersionMode, publish2.VersionMultipleMode).Set(publish2.ScheduleMode, publish2.ModeOff).
			Where("id = ? AND status = ?", version.GetId(), Status_published).First(iRecord)
		liveRecord := iRecord.(QorMicroSiteInterface)
		if liveRecord.GetId() != 0 {
			for _, o := range liveRecord.GetFilesPathWithSiteURL() {
				oss.Storage.Delete(o)
			}

			liveRecord.SetStatus(Status_unpublished)
			if err1 = tx.Save(liveRecord).Error; err1 != nil {
				return
			}

			if liveRecord.GetVersionName() != version.GetVersionName() {
				if err1 = liveRecord.UnPublishCallBack(_db, liveRecord.GetMicroSiteURL()); err1 != nil {
					return
				}
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

		if liveRecord.GetVersionName() != version.GetVersionName() {
			err1 = version.PublishCallBack(_db, version.GetMicroSiteURL())
		}
		return
	})

	return
}
