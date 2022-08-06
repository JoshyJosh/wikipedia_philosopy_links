package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

const BASE_URL = "https://en.wikipedia.org"

var hopCount int
var visitedPages map[string]struct{} = map[string]struct{}{}

func main() {
	// @todo set custom input url
	url := fmt.Sprintf("%s/wiki/Wikipedia:Getting_to_Philosophy", BASE_URL)

	wikiPage, err := http.Get(url)
	if err != nil {
		panic(errors.Wrapf(err, "failed to get wiki page: %s", url))
	}

	pageNodes, err := html.Parse(wikiPage.Body)
	if err != nil {
		panic(errors.Wrapf(err, "failed to parse wiki page: %s", url))
	}

	findAndFollowLink(pageNodes)

	fmt.Println("stats")

	fmt.Printf("Hop count: %d", hopCount)
}

func followLink(url string) *html.Node {
	url = BASE_URL + url
	fmt.Println(url)
	_, ok := visitedPages[url]
	if ok {
		panic(fmt.Sprintf("page already visited: %s", url))
	}

	visitedPages[url] = struct{}{}

	wikiPage, err := http.Get(url)
	if err != nil {
		panic(errors.Wrapf(err, "failed to get wiki page: %s", url))
	}
	hopCount++

	pageNodes, err := html.Parse(wikiPage.Body)
	if err != nil {
		panic(errors.Wrapf(err, "failed to parse wiki page: %s", url))
	}

	return pageNodes
}

func parseList(n *html.Node) string {
	for s := n.FirstChild; n != nil; s = s.NextSibling {
		if s.Type != html.ElementNode && s.Data != "li" {
			panic("unexpected non li node")
		}

		for s2 := s.FirstChild; s != nil; s2 = s2.NextSibling {
			if s2.Type != html.ElementNode || s2.Data != "a" {
				continue
			}

			for _, attr := range s2.Attr {
				if attr.Key == "href" {
					fmt.Printf("following first link: %s\n", attr.Val)

					return attr.Val
				}
			}
		}
	}

	return ""
}

func findFirstLink(n *html.Node) string {
	for s := n.FirstChild; s != nil; s = s.NextSibling {
		if s.Type == html.ElementNode && s.Data == "ul" {
			link := parseList(s)
			if link == "" {
				continue
			}

			fmt.Printf("following first link: %s\n", link)

			return link
		}

		if s.Type != html.ElementNode || s.Data != "p" {
			continue
		}

		for s2 := s.FirstChild; s2 != nil; s2 = s2.NextSibling {
			if s2.Type != html.ElementNode || s2.Data != "a" {
				continue
			}

			for _, attr := range s2.Attr {
				if attr.Key == "href" {
					fmt.Printf("following first link: %s\n", attr.Val)

					return attr.Val
				}
			}
		}
	}

	panic(fmt.Sprintf("failed to find link in node: %#v\n", n))
}

func findAndFollowLink(n *html.Node) {
	body := findArticleBody(n)
	if body == nil {
		panic("failed to find article body")
	}

	link := findFirstLink(body)
	if strings.Contains(link, "/wiki/Philosophy") {
		fmt.Printf("found philosophy link: %s\n", link)
		return
	}

	findAndFollowLink(followLink(link))
}

func findArticleBody(n *html.Node) *html.Node {
	for i := range n.Attr {
		if n.Attr[i].Key == "class" && n.Attr[i].Val == "mw-parser-output" {
			fmt.Println("found article body")
			fmt.Printf("%#v\n", n)

			return n
		}
	}

	var wg sync.WaitGroup
	var body *html.Node
childLoop:
	for s := n.FirstChild; s != nil; s = s.NextSibling {

		for i := range s.Attr {
			if s.Attr[i].Key == "class" && s.Attr[i].Val == "mw-indicators" {
				continue childLoop
			}
		}

		wg.Add(1)
		go func(n *html.Node, wg *sync.WaitGroup) {
			defer wg.Done()

			// skip traversal if a body was already found
			if body != nil {
				return
			}

			b := findArticleBody(n)
			if b != nil {
				body = b
			}
		}(s, &wg)
	}
	wg.Wait()

	return body
}
