package microsite

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/media/oss"
	"github.com/theplant/gormutils"
)

func Unpublish(db *gorm.DB, version QorMicroSiteInterface, arg *admin.ActionArgument) (err error) {
	tableName := db.NewScope(version).TableName()

	err = gormutils.Transact(db, func(tx *gorm.DB) (err1 error) {
		defer func() {
			if err1 != nil {
				eventType := fmt.Sprintf("%s:UnpublishingError", strings.Title(tableName))
				fmt.Printf("%v, error: %v\n", eventType, err1.Error())
			}
		}()
		now := gorm.NowFunc()
		version.SetStatus(Status_unpublished)
		version.SetScheduledEndAt(&now)
		if err1 = tx.Save(version).Error; err1 != nil {
			return
		}

		if err1 = version.UnPublishCallBack(tx, version.GetMicroSiteURL(), arg); err1 != nil {
			return
		}

		if s3, ok := oss.Storage.(DeleteObjecter); ok {
			err1 = s3.DeleteObjects(version.GetFilesPathWithSiteURL())
		} else {
			for _, o := range version.GetFilesPathWithSiteURL() {
				oss.Storage.Delete(o)
			}
		}

		return
	})

	return
}
