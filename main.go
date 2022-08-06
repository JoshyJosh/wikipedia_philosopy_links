package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

const BASE_URL = "https://en.wikipedia.org"

var hopCount int
var visitedPages map[string]struct{} = map[string]struct{}{}

func main() {
	// @todo set custom input url
	url := fmt.Sprintf("%s/wiki/Wikipedia:Getting_to_Philosophy", BASE_URL)
	// url := fmt.Sprintf("%s/wiki/Reality", BASE_URL)

	wikiPage, err := http.Get(url)
	if err != nil {
		logrus.Error(errors.Wrapf(err, "failed to get wiki page: %s", url))
		return
	}

	pageNodes, err := html.Parse(wikiPage.Body)
	if err != nil {
		logrus.Error(errors.Wrapf(err, "failed to parse wiki page: %s", url))
		return
	}

	err = findAndFollowLink(pageNodes)
	if err != nil {
		logrus.Error(errors.Wrap(err, "failed to find Philosophy page"))
	}

	logrus.Info("stats")
	logrus.Infof("Hop count: %d", hopCount)
}

func followLink(url string) (*html.Node, error) {
	url = BASE_URL + url
	logrus.Info(url)

	if _, ok := visitedPages[url]; ok {
		return nil, fmt.Errorf("page already visited: %s", url)
	}

	visitedPages[url] = struct{}{}

	wikiPage, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get wiki page: %s", url)
	}
	hopCount++

	pageNodes, err := html.Parse(wikiPage.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse wiki page: %s", url)
	}

	return pageNodes, nil
}

func parseList(n *html.Node) (string, error) {
	for s := n.FirstChild; n != nil; s = s.NextSibling {
		if s.Type != html.ElementNode && s.Data != "li" {
			return "", errors.New("unexpected non li node")
		}

		for s2 := s.FirstChild; s != nil; s2 = s2.NextSibling {
			if s2.Type != html.ElementNode || s2.Data != "a" {
				continue
			}

			for _, attr := range s2.Attr {
				if attr.Key == "href" {
					logrus.Infof("following first link: %s\n", attr.Val)

					return attr.Val, nil
				}
			}
		}
	}

	return "", nil
}

func findFirstLink(n *html.Node) (string, error) {
	for s := n.FirstChild; s != nil; s = s.NextSibling {
		if s.Type == html.ElementNode && s.Data == "ul" {
			link, err := parseList(s)
			if err != nil {
				return "", errors.Wrap(err, "failed to parse unsorted list")
			}

			if link == "" {
				continue
			}

			logrus.Infof("following first link: %s\n", link)

			return link, nil
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
					logrus.Infof("following first link: %s\n", attr.Val)

					return attr.Val, nil
				}
			}
		}
	}

	return "", fmt.Errorf("failed to find link in node: %#v\n", n)
}

func findAndFollowLink(n *html.Node) error {
	body := findArticleBody(n)
	if body == nil {
		return errors.New("failed to find article body")
	}

	link, err := findFirstLink(body)
	if err != nil {
		return errors.Wrap(err, "failed to find first link")
	}

	if strings.Contains(link, "/wiki/Philosophy") {
		logrus.Infof("found philosophy link: %s\n", link)
		return nil
	}

	linkNode, err := followLink(link)
	if err != nil {
		return errors.Wrap(err, "failed to follow link")
	}

	err = findAndFollowLink(linkNode)
	if err != nil {
		return errors.Wrap(err, "failed to find and follow link")
	}

	return nil
}

func findArticleBody(n *html.Node) *html.Node {
	for i := range n.Attr {
		if n.Attr[i].Key == "class" && n.Attr[i].Val == "mw-parser-output" {
			logrus.Info("found article body")
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
