# Wikipedia Philosophy links

This is an experiment to check the "Everything links to philosophy" claim:
https://en.wikipedia.org/wiki/Wikipedia:Getting_to_Philosophy

*TLDR* the algorithm works like this:
1. Clicking on the first non-parenthesized, non-italicized link
2. Ignoring external links, links to the current page, or red links (links to non-existent pages)
3. Stopping when reaching "Philosophy", a page with no links or a page that does not exist, or when a loop occurs

## Test results

The links loop usually on the "Branches of science" page. For reference I double checked with [xefers wikipedia](https://www.xefer.com/wikipedia) project. Seemingly the only difference is doing the API calls instead of parsing the html pages, however using the random button i found that false positives can still happend (try typing in `Crichton, Alabama,Gu`). That being said I love the project overall and its a nice visual representation.

As far as I can tell the `Getting to Philosophy` phenomenon is false, though it might have still worked in the past.