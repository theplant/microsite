package microsite

import (
	"fmt"

	"github.com/qor/admin"
)

func ChangeStatus(argument *admin.ActionArgument, action string) (err error) {
	db := argument.Context.DB

	for _, record := range argument.FindSelectedRecords() {
		if version, ok := record.(QorMicroSiteInterface); ok {
			argument.Context.Result = version
			status := changeStatusActionMap[action]
			if action == Action_publish {
				if err = Publish(db, version, argument); err != nil {
					return
				}
			} else if action == Action_republish {
				if err = Republish(db, version, argument); err != nil {
					return
				}
			} else if action == Action_unpublish {
				if err = Unpublish(db, version, argument); err != nil {
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
