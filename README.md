# Microsite

## Usage

Initialize microsite with a S3 instance, a QOR admin instance, `microsite.QorMicroSite` struct and an optional `admin.Config`.

```go
microsite.Init(s3, adm, &microsite.QorMicroSite{}, &admin.Config{Name: "Microsites"})
```

If you need to customize `QorMicroSite` which is mandatory for handle microsite url callback(add or remove to/from sitemap etc.). Define your own struct like this

```go
type QorMicroSite struct {
	microsite.QorMicroSite
}
```

and invoke the `microsite.Init` function like this

```go
microsite.Init(s3, adm, &QorMicroSite{}, &admin.Config{Name: "Microsites"})
```

Then you can implement 2 functions to handle the url with sitemap like this. If an error occurred, the publish or unpublish action will be reverted.

```go
func (site QorMicroSite) PublishCallBack(db *gorm.DB, siteURL string) error {
    // siteURL is the microsite URL online, you can add it to your sitemap according to your own business logic
    // HINT: you can validate the uniqueness of the url in the sitemap to avoid microsite overwrite the main path in your site.
	return nil
}

func (site QorMicroSite) UnPublishCallBack(db *gorm.DB, siteURL string) error {
    // siteURL is the URL of the microsite that about to unpublished. you can remove it from your sitemap
	return nil
}
```


## To get the ready for publish or unpublish sites.

Since this microsite is designed for the independent frontend&backend infrastructure. so the old publish2 is no longer efficient. We have to explicitly trigger the "publish pages to s3" operation by a cronjob. So we made 2 functions to retrieve the sites that to be published or unpublished. so that your cronjob could invoke and do the publishing and unpublishing.

### For publishing
Use `TobePublishedMicrosites` function. the last parameter is the status your project treat as "ready for publish". e.g. "Approved".

### For unpublishing
Use `TobeUnpublishedMicrosite` function. the last parameter is the status your project treat as "ready for unpublish". usually it is "Published".

