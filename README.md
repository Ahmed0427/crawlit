# crawlit

A Golang CLI application that generates an `internal links` report for any
website on the internet by crawling each page of the site.

### Quickstrat
```console
go build 
./crawlit <URL> <MAX_PAGES> <MAX_GOROUTINES>
```
be careful when setting the MAX_GOROUTINES arg not do more than 200 or 300
