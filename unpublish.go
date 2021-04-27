package microsite

import (
	"context"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/oss"
	"github.com/theplant/gormutils"
)

func Unpublish(ctx context.Context, version QorMicroSiteInterface, printActivityLog bool) (err error) {
	_db := ctx.Value("DB").(*gorm.DB)
	tableName := _db.NewScope(version).TableName()

	err = gormutils.Transact(_db, func(tx *gorm.DB) (err1 error) {
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

		for _, o := range version.GetFilesPathWithSiteURL() {
			oss.Storage.Delete(o)
		}

		return version.UnPublishCallBack(_db, version.GetMicroSiteURL())

	})

	return
}
