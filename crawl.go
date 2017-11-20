package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"

	"golang.org/x/net/html"
)

// Link is a recursive type for storing links, parents of links and children of links
type Link struct {
	parent   *Link
	value    string
	children []*Link
}

// addChild appends Link pointer to the children of a parent Link
func (l *Link) addChild(v *Link) {
	l.children = append(l.children, v)
}

// Implements the string interface for how links should be displayed
func (l Link) String() string {
	return l.value
}

// Fetch makes a HTTP GET request for resource given its domain and path
// Converts the HTML into a tokenised object where all valid links are extracted
// It returns map containing all of the unique internal links found on the webpage
func fetch(domain *url.URL, path string) []string {
	response, err := http.Get(domain.String() + path)

	fmt.Printf("\rFetching: %s", path+"                                     ")

	if err != nil {
		panic(err)
	}

	defer response.Body.Close() // Close the response after function completed

	z := html.NewTokenizer(response.Body) // Tokenise the body of the html
	allUrls := make(map[string]bool)

	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken: // Terminal token
			return getKeys(allUrls)
		case tt == html.StartTagToken:
			t := z.Token()
			isAnchor := t.Data == "a"
			if isAnchor {
				dataAttr := t.Attr
				for _, v := range dataAttr {
					// Make sure we are dealing with links to pages
					if v.Key == "href" {
						u, err := url.Parse(v.Val)
						if err != nil {
							break
						} else {
							// Only use links that are related to the host or relative
							if (u.Host == domain.Host || u.Host == "") && (string(u.Path) != "") {
								firstChar := string(u.Path[0])
								if firstChar == "/" {
									allUrls[strip(u.Path)] = true
								}
								if firstChar == "." {
									// dont include relative links
								}
								if firstChar != "/" && firstChar != "." {
									allUrls[strip(path+"/"+u.Path)] = true
								}
							}
						}
					}
				}
			}
		}
	}
}

// Crawl traverses a website given its domain name and path
// Conducts a Breadth First Search finding links and constructing a tree/graph
// Returns a pointer to the root of the tree/graph of all connected urls
func crawl(domain string, path string) *Link {
	// Parse the link into a url.URL struct
	domainObj, err := url.Parse(domain)

	if err != nil {
		panic(err)
	}

	visitedPages := map[string]*Link{} // Pointers to nodes visited in the traversal
	root := &Link{nil, path, []*Link{}}
	visitedPages[path] = root
	queue := []*Link{}
	queue = append(queue, root)

	for len(queue) > 0 {
		link := queue[len(queue)-1]              // Get the first link in the queue
		queue = queue[:len(queue)-1]             // Pop the link from the queue
		children := fetch(domainObj, link.value) // Fetch all the children from the link

		for _, v := range children {
			if _, ok := visitedPages[v]; !ok {
				child := &Link{link, v, []*Link{}}
				queue = append(queue, child) // Add the unseen link to tthe queue
				visitedPages[v] = child      // Mark the link as visted
				link.addChild(child)         // Add the link to the graph/tree
			} else {
				// link.addChild(l) // If we want a graph add l instead of _ on line 108
			}
		}
	}
	return root // return a graph of connected websites
}

// Strip removes all "/" from the end of links
func strip(s string) string {
	strLen := len(s)
	if strLen > 1 {
		if string(s[strLen-1]) == "/" {
			return s[:strLen-1]
		}
	}
	return s
}

// getKeys returns all of the keys from map
func getKeys(m map[string]bool) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// printSitemap prints to the console a sitemap of a website given a *Link
func printSitemap(l *Link, indent string) {
	if len(l.children) == 0 {
		return
	}
	for _, v := range l.children {
		fmt.Println(" " + indent + v.value)
		printSitemap(v, indent+"   ")
	}
}

func main() {
	domain := os.Args[1]                  // Get the first command line argument
	root := crawl("https://"+domain, "/") // crawl the given domain
	fmt.Println("Fetching Completed\n")
	fmt.Println(domain + " sitemap" + "\n/")
	printSitemap(root, "") // Print the sitemap
}
