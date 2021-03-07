# Scrapen
Scrapen is a tool to download article content from web pages.

Features include:

- following redirects
- extract the main content (article) from the web site
- clean up the resulting HTML
- download referenced images
- extract additional metadata

## Status
Early development, but should be usable.

## Usage

```go
package main

import (
    "fmt"
    "github.com/akeil/scrapen"
)

func main() {
    url := "https://golang.org/doc/effective_go"

    article, err := scrapen.Scrape(url, nil)
    if err != nil {
        fmt.Printf("Scrape failed: %v\n", err)
        return
    }

    fmt.Printf("Title: %v\n", article.Title)
    fmt.Printf("Site: %v\n", article.Site)
    fmt.Printf("Final URL: %v\n", article.ActualURL)
    fmt.Println("HTML:")
    fmt.Println(article.HTML)
}
```

The behavior of a scraping task can be controlled with the `Options` struct:

```go

o := &scrapen.Options{
    Metadata:       true,
    Readability:    true,
    Clean:          true,
    DownloadImages: true,
    Store:          MyStoreImplementation(),
}
article, err := scrapen.Scrape(url, o)
```

## CLI
A small command line tool is included.

**Usage**

```
$ scrapen https://golang.org/doc/effective_go
```

Will write the resulting HTML page to a local file `./output.html`.
