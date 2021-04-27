package microsite

import (
	"context"
	"fmt"

	"github.com/qor/admin"
)

func ChangeStatus(argument *admin.ActionArgument, action string) (err error) {
	ctx := argument.Context.Request.Context()
	db := argument.Context.DB
	ctx = context.WithValue(ctx, "DB", db)
	for _, record := range argument.FindSelectedRecords() {
		if version, ok := record.(QorMicroSiteInterface); ok {
			argument.Context.Result = version
			status := changeStatusActionMap[action]

			if action == Action_publish {
				if err = Publish(ctx, version, true); err != nil {
					return
				}
			} else if action == Action_republish {
				if err = Republish(ctx, version, true); err != nil {
					return
				}
			} else if action == Action_unpublish {
				if err = Unpublish(ctx, version, true); err != nil {
					return
				}
			} else {
				version.SetStatus(status)
				if err = db.Save(version).Error; err != nil {
					return
				}
			}
		} else {
			return fmt.Errorf("record is not completely implement QorMicroSiteInterface")
		}
	}
	return
}
