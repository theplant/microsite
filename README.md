# microsite

# To get the ready for publish or unpublish sites.

Since this microsite is designed for the independent frontend&backend infrastructure. so the old publish2 is no longer efficient. We have to explicitly trigger the "publish pages to s3" operation by a cronjob. So we made 2 functions to retrieve the sites that to be published or unpublished. so that your cronjob could invoke and do the publishing and unpublishing.

## For publishing
Use `TobePublishedMicrosites` function. the last parameter is the status your project treat as "ready for publish". e.g. "Approved".

## For unpublishing
Use `TobeUnpublishedMicrosite` function. the last parameter is the status your project treat as "ready for unpublish". usually it is "Published".

