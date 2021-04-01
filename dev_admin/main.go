package main

import (
	"net/http"

	"github.com/qor/media/oss"
	"github.com/qor/oss/s3"
	"github.com/qor/publish2"

	"github.com/jinzhu/configor"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/qor/admin"
	"github.com/qor/media"
	"github.com/qor/qor"
	"github.com/qor/validations"
	appkitdb "github.com/theplant/appkit/db"
	appkitlog "github.com/theplant/appkit/log"
	"github.com/theplant/appkit/server"
	"github.com/theplant/microsite"
)

func main() {
	loger := appkitlog.Default()
	var err error

	if err != nil {
		panic(err)
	}

	cfg := appkitdb.Config{}

	err = configor.New(&configor.Config{ENVPrefix: "M"}).Load(&cfg)
	if err != nil {
		panic(err)
	}

	var s3cfg s3.Config
	if err := configor.New(&configor.Config{ENVPrefix: "M_S3"}).Load(&s3cfg); err == nil && s3cfg.Bucket != "" {
		oss.Storage = s3.New(&s3cfg)
	}

	var db *gorm.DB
	db, err = appkitdb.New(loger, cfg)
	if err != nil {
		panic(err)
	}
	validations.RegisterCallbacks(db)
	publish2.RegisterCallbacks(db)
	media.RegisterCallbacks(db)

	if err != nil {
		panic(err)
	}

	adm := admin.New(&qor.Config{DB: db})
	adm.SetSiteName("Sample Admin")
	microsite.New(adm, oss.Storage, oss.Storage)
	mux := http.DefaultServeMux
	adm.MountTo("/admin", mux)
	server.ListenAndServe(server.Config{Addr: ":7000"}, loger, mux)

}
