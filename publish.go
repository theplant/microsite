package microsite

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/oss"
	"github.com/qor/publish2"
	"github.com/theplant/gormutils"
)

func Publish(db *gorm.DB, version QorMicroSiteInterface, printActivityLog bool) (err error) {
	tableName := db.NewScope(version).TableName()

	err = gormutils.Transact(db, func(tx *gorm.DB) (err1 error) {
		defer func() {
			if err1 != nil {
				eventType := fmt.Sprintf("%s:PublishingError", strings.Title(tableName))
				fmt.Printf("%v, error: %v\n", eventType, err1.Error())
			}
		}()

		// Find possible online version
		iRecord := reflect.New(reflect.TypeOf(version).Elem()).Interface()
		if err1 = admDB.Set(publish2.VersionMode, publish2.VersionMultipleMode).Set(publish2.ScheduleMode, publish2.ModeOff).
			Where("id = ? AND status = ?", version.GetId(), Status_published).Where("version_name <> ?", version.GetVersionName()).
			First(iRecord).Error; err1 != nil {
			return
		}

		liveRecord, ok := iRecord.(QorMicroSiteInterface)
		if !ok {
			return errors.New("given record doesn't implement QorMicroSiteInterface")
		}

		// If there is a published version, unpublish it
		if liveRecord.GetId() != 0 {
			for _, o := range liveRecord.GetFilesPathWithSiteURL() {
				oss.Storage.Delete(o)
			}

			liveRecord.SetStatus(Status_unpublished)
			if err1 = tx.Save(liveRecord).Error; err1 != nil {
				return
			}

			if err1 = liveRecord.UnPublishCallBack(db, liveRecord.GetMicroSiteURL()); err1 != nil {
				return
			}
		}

		// Publish given version
		version.SetStatus(Status_published)
		version.SetVersionPriority(fmt.Sprintf("%v", time.Now().UTC().Format(time.RFC3339)))
		if err1 = tx.Save(version).Error; err1 != nil {
			return
		}

		if _, err1 = UnzipPkgAndUpload(version.GetMicroSitePackage().Url, version.GetMicroSiteURL()); err1 != nil {
			return
		}

		return version.PublishCallBack(db, version.GetMicroSiteURL())
	})

	return
}
