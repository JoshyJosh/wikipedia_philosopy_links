package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

func main() {
	// @todo set custom input url
	url := "https://en.wikipedia.org/wiki/Wikipedia:Getting_to_Philosophy"

	wikiPage, err := http.Get(url)
	if err != nil {
		panic(errors.Wrapf(err, "failed to get wiki page: %s", url))
	}

	pageNodes, err := html.Parse(wikiPage.Body)
	if err != nil {
		panic(errors.Wrapf(err, "failed to parse wiki page: %s", url))
	}

	findArticleBody(pageNodes)
}

// @todo find first link
// func findFirstLink() {

// }

func findArticleBody(n *html.Node) {
	for _, attr := range n.Attr {
		if attr.Key == "class" && attr.Val == "mw-parser-output" {
			fmt.Println("found article body")
			fmt.Printf("%#v", n)

			return
		}
	}

	var wg sync.WaitGroup
	for s := n.FirstChild; s != nil; s = s.NextSibling {
		wg.Add(1)
		go func(n *html.Node, wg *sync.WaitGroup) {
			defer wg.Done()
			findArticleBody(n)
		}(s, &wg)
	}
	wg.Wait()
}
