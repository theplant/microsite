package microsite_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/qor/publish2"
	"github.com/qor/qor/test/utils"
	"github.com/theplant/microsite"
)

func TestTobeOnlineSites(t *testing.T) {
	db := utils.PrepareDBAndTables(&microsite.QorMicroSite{})
	publish2.RegisterCallbacks(db)

	now := time.Now()
	m1 := NewMicrosite(microsite.Status_draft)
	m2 := NewMicrosite(microsite.Status_draft)

	aWhileAgo := now.Add(-1 * time.Minute)
	aWhileLater := now.Add(1 * time.Minute)

	m1.SetScheduledStartAt(&aWhileAgo)
	m2.SetScheduledStartAt(&aWhileLater)

	utils.AssertNoErr(t, db.Create(&m1).Error)
	utils.AssertNoErr(t, db.Create(&m2).Error)

	sites, err := microsite.GetSites(db, microsite.Status_draft)
	if err != nil {
		t.Fatal(err)
	}

	if len(sites) != 1 {
		t.Error("not returning correct numbers of sites")
	}
	if sites[0].GetId() != m1.ID {
		t.Error("not returning correct site")
	}
}

func TestTobeOfflineSites(t *testing.T) {
	db := utils.PrepareDBAndTables(&microsite.QorMicroSite{})
	publish2.RegisterCallbacks(db)

	now := time.Now()
	m1 := NewMicrosite(microsite.Status_published)
	m2 := NewMicrosite(microsite.Status_published)

	aWhileAgo := now.Add(-1 * time.Minute)
	aWhileLater := now.Add(1 * time.Minute)

	m1.SetScheduledEndAt(&aWhileLater)
	m2.SetScheduledEndAt(&aWhileAgo)

	utils.AssertNoErr(t, db.Create(&m1).Error)
	utils.AssertNoErr(t, db.Create(&m2).Error)

	sites, err := microsite.GetSites(db, microsite.Status_draft)
	if err != nil {
		t.Fatal(err)
	}

	if len(sites) != 1 {
		t.Error("not returning correct numbers of sites")
	}
	if sites[0].GetId() != m1.ID {
		t.Error("not returning correct site")
	}
}

func NewMicrosite(status string) microsite.QorMicroSite {
	return microsite.QorMicroSite{Name: randomdata.SillyName(), URL: fmt.Sprintf("/%s", randomdata.SillyName()), Status: status}
}
