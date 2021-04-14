package microsite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/oss"
	"github.com/theplant/appkit/db"
	"github.com/theplant/gormutils"
)

func Unpublish(ctx context.Context, version QorMicroSiteInterface, printActivityLog bool) (err error) {
	_db := db.MustGetGorm(ctx)
	tableName := _db.NewScope(version).TableName()

	err = gormutils.Transact(_db, func(tx *gorm.DB) (err1 error) {
		defer func() {
			if err1 != nil {
				eventType := fmt.Sprintf("%s:UnpublishingError", strings.Title(tableName))
				fmt.Printf("%v, error: %v\n", eventType, err1.Error())
			}
		}()

		version.SetStatus(Status_unpublished)
		version.SetVersionPriority(fmt.Sprintf("%v", version.GetCreatedAt().UTC().Format(time.RFC3339)))
		if err1 = tx.Save(version).Error; err1 != nil {
			return
		}

		if err1 = version.SitemapHandler(_db, version.GetMicroSiteURL(), Action_unpublish); err1 != nil {
			return
		}

		objs, err1 := oss.Storage.List(version.GetMicroSiteURL())
		for _, o := range objs {
			oss.Storage.Delete(o.Path)
		}

		return

	})

	return
}
