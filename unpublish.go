package microsite

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/oss"
	"github.com/theplant/gormutils"
)

func Unpublish(db *gorm.DB, version QorMicroSiteInterface, printActivityLog bool) (err error) {
	tableName := db.NewScope(version).TableName()

	err = gormutils.Transact(db, func(tx *gorm.DB) (err1 error) {
		defer func() {
			if err1 != nil {
				eventType := fmt.Sprintf("%s:UnpublishingError", strings.Title(tableName))
				fmt.Printf("%v, error: %v\n", eventType, err1.Error())
			}
		}()

		version.SetStatus(Status_unpublished)
		if err1 = tx.Save(version).Error; err1 != nil {
			return
		}

		if err1 = version.UnPublishCallBack(tx, version.GetMicroSiteURL()); err1 != nil {
			return
		}

		for _, o := range version.GetFilesPathWithSiteURL() {
			oss.Storage.Delete(o)
		}

		return
	})

	return
}
