package microsite

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
)

func ChangeStatus(argument *admin.ActionArgument, action string) (err error) {
	db := argument.Context.DB

	for _, record := range argument.FindSelectedRecords() {
		if version, ok := record.(QorMicroSiteInterface); ok {
			argument.Context.Result = version
			status := changeStatusActionMap[action]
			now := gorm.NowFunc()

			if action == Action_publish {
				if err = db.Model(version).UpdateColumns(map[string]interface{}{"scheduled_start_at": &now}).Error; err != nil {
					return
				}
				if err = Publish(db, version, true); err != nil {
					return
				}
			} else if action == Action_republish {
				//have to reset scheduled_end_at
				if err = db.Model(version).UpdateColumns(map[string]interface{}{"scheduled_start_at": &now, "scheduled_end_at": nil}).Error; err != nil {
					return
				}

				if err = Republish(db, version, true); err != nil {
					return
				}
			} else if action == Action_unpublish {
				if err = db.Model(version).UpdateColumns(map[string]interface{}{"scheduled_end_at": &now}).Error; err != nil {
					return
				}

				if err = Unpublish(db, version, true); err != nil {
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
