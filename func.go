package microsite

import (
	"fmt"
	"html/template"

	"github.com/qor/admin"
	"github.com/qor/publish2"
	"github.com/qor/roles"
)

func getVersions(context *admin.Context) template.HTML {
	records := context.Resource.NewSlice()
	record := context.Resource.NewStruct()
	primaryQuerySQL, primaryParams := context.Resource.ToPrimaryQueryParams(context.ResourceID, context.Context)
	tx := context.GetDB().Set(admin.DisableCompositePrimaryKeyMode, "on").Set(publish2.VersionMode, publish2.VersionMultipleMode).Set(publish2.ScheduleMode, publish2.ModeOff).Set(publish2.VisibleMode, publish2.ModeOff)
	tx.Where(primaryQuerySQL, primaryParams...).First(record)

	scope := tx.NewScope(record)
	tx.Find(records, fmt.Sprintf("%v = ?", scope.PrimaryKey()), scope.PrimaryKeyValue())
	return context.Funcs(template.FuncMap{
		"version_metas": func() (metas []*admin.Meta) {

			for _, name := range []string{"VersionName", "ScheduledStartAt", "ScheduledEndAt", "PublishReady", "PublishLiveNow"} {
				if meta := context.Resource.GetMeta(name); meta != nil && meta.HasPermission(roles.Read, context.Context) {
					metas = append(metas, meta)
				}
			}
			return
		},
	}).Render("dashboard", records)

}
