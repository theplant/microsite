package microsite

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/media/oss"
	"github.com/qor/publish2"
	"github.com/theplant/gormutils"
)

func Republish(db *gorm.DB, version QorMicroSiteInterface, printActivityLog bool) (err error) {
	tableName := db.NewScope(version).TableName()

	err = gormutils.Transact(db, func(tx *gorm.DB) (err1 error) {
		defer func() {
			if err1 != nil {
				eventType := fmt.Sprintf("%s:RepublishingError", strings.Title(tableName))
				fmt.Printf("%v, error: %v\n", eventType, err1.Error())
			}
		}()

		iRecord := reflect.New(reflect.TypeOf(version).Elem()).Interface()
		if err1 = db.Set(admin.DisableCompositePrimaryKeyMode, "on").Set(publish2.VersionMode, publish2.VersionMultipleMode).Set(publish2.ScheduleMode, publish2.ModeOff).
			Where("id = ? AND status = ?", version.GetId(), Status_published).First(iRecord).Error; err1 != nil && err1 != gorm.ErrRecordNotFound {
			return
		}

		liveRecord, ok := iRecord.(QorMicroSiteInterface)
		if !ok {
			return errors.New("given record doesn't implement QorMicroSiteInterface")
		}
		now := gorm.NowFunc()
		if liveRecord.GetId() != 0 {
			if s3, ok := oss.Storage.(DeleteObjecter); ok {
				err1 = s3.DeleteObjects(liveRecord.GetFilesPathWithSiteURL())
			} else {
				for _, o := range liveRecord.GetFilesPathWithSiteURL() {
					oss.Storage.Delete(o)
				}
			}
			if err1 != nil {
				return
			}

			liveRecord.SetStatus(Status_unpublished)
			liveRecord.SetScheduledEndAt(&now)
			if err1 = tx.Save(liveRecord).Error; err1 != nil {
				return
			}

			if liveRecord.GetVersionName() != version.GetVersionName() {
				if err1 = liveRecord.UnPublishCallBack(db, liveRecord.GetMicroSiteURL()); err1 != nil {
					return
				}
			}
		}

		version.SetStatus(Status_published)
		version.SetScheduledStartAt(&now)
		version.SetScheduledEndAt(nil)
		version.SetVersionPriority(fmt.Sprintf("%v", now.UTC().Format(time.RFC3339)))
		if err1 = tx.Save(version).Error; err1 != nil {
			return
		}

		if _, err1 = UnzipPkgAndUpload(version.GetMicroSitePackage().Url, version.GetMicroSiteURL()); err1 != nil {
			return
		}

		if liveRecord.GetVersionName() != version.GetVersionName() {
			err1 = version.PublishCallBack(db, version.GetMicroSiteURL())
		}
		return
	})

	return
}
