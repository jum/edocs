package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/jum/htmlelements"
	"golang.org/x/net/html"
)

const base = "https://www.ndr.de/fernsehen/sendungen/die-ernaehrungsdocs/rezepte/index.html"
const debug = false

type recipe struct {
	link  string
	title string
}

type category struct {
	link    string
	title   string
	recipes []recipe
}

var (
	outdir        = flag.String("outdir", "outdir", "output directory")
	compressSpace = regexp.MustCompile(`\s+`)
)

func main() {
	flag.Parse()
	u, err := url.Parse(base)
	if err != nil {
		panic(err)
	}
	var categories []category
	markDown := make(map[string]string)
	images := make(map[string]string)
	err = os.MkdirAll(*outdir, 0700)
	if err != nil {
		panic(err)
	}
	var f io.ReadCloser
	if debug {
		f, err = os.Open("index.html")
		if err != nil {
			panic(err)
		}
	} else {
		resp, err := http.Get(base)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode/100 != 2 {
			fmt.Fprintf(os.Stderr, "%s: %s\n", base, resp.Status)
			os.Exit(1)
		}
		f = resp.Body
	}
	doc, err := html.Parse(f)
	if err != nil {
		panic(err)
	}
	f.Close()
	sections := htmlelements.GetElementsByTagName(doc, "section")
	for _, a := range htmlelements.GetElementsByTagName(sections[0], "a") {
		title := htmlelements.GetAttribute(a, "title")
		if strings.HasPrefix(title, "Zum Artikel: ") {
			title = title[13:]
		}
		title = strings.TrimSpace(title)
		u.Path = htmlelements.GetAttribute(a, "href")
		categories = append(categories, category{link: u.String(), title: title})
	}
	fmt.Printf("%d Kategorien\n", len(categories))
	for i := range categories {
		var nextPage = categories[i].link
		for len(nextPage) > 0 {
			var f io.ReadCloser
			if debug {
				f, err = os.Open("rezeptdb242.html")
				if err != nil {
					panic(err)
				}
			} else {
				resp, err := http.Get(nextPage)
				if err != nil {
					panic(err)
				}
				if resp.StatusCode/100 != 2 {
					fmt.Fprintf(os.Stderr, "%s: %s\n", nextPage, resp.Status)
					os.Exit(1)
				}
				f = resp.Body
			}
			doc, err = html.Parse(f)
			if err != nil {
				panic(err)
			}
			f.Close()
			nodes := htmlelements.GetElementsByClassName(doc, "rezepteText")
			for _, n := range nodes {
				a := htmlelements.GetElementsByTagName(n, "a")
				title := nodeText(a[0])
				title = strings.ReplaceAll(title, "\n", " ")
				title = compressSpace.ReplaceAllString(title, " ")
				title = strings.TrimSpace(title)
				u.Path = htmlelements.GetAttribute(a[0], "href")
				link := u.String()
				categories[i].recipes = append(categories[i].recipes, recipe{title: title, link: link})
			}
			nodes = htmlelements.GetElementsByClassName(doc, "pagination")
			nextPage = ""
			for _, a := range htmlelements.GetElementsByTagName(nodes[0], "a") {
				if htmlelements.GetAttribute(a, "title") == "weiter" {
					u.Path = htmlelements.GetAttribute(a, "href")
					nextPage = u.String()
				}
			}
			if debug {
				break
			}
		}
		if debug {
			break
		}
	}
	if debug {
		spew.Dump(categories)
	}
	for _, c := range categories {
		fmt.Printf("Lese %s\n", c.title)
		for _, r := range c.recipes {
			if _, ok := markDown[r.link]; ok {
				continue
			}
			var f io.ReadCloser
			if debug {
				f, err = os.Open("rezeptdb6_id-9643_broadcast-1530_station-ndrtv.html")
				if err != nil {
					panic(err)
				}
			} else {
				resp, err := http.Get(r.link)
				if err != nil {
					panic(err)
				}
				if resp.StatusCode/100 != 2 {
					fmt.Fprintf(os.Stderr, "%s: %s\n", r.link, resp.Status)
					os.Exit(1)
				}
				f = resp.Body
			}
			doc, err := html.Parse(f)
			if err != nil {
				panic(err)
			}
			f.Close()
			recipeNode := htmlelements.GetElementByID(doc, "rezepte")
			receipeImg, recipeText := recipe2md(recipeNode)
			markDown[r.link] = recipeText
			images[r.link] = receipeImg
			if debug {
				break
			}
		}
		if debug {
			break
		}
	}
	if debug {
		spew.Dump(markDown)
		spew.Dump(images)
	}
	for _, c := range categories {
		fmt.Printf("Lese Rezepte für %s\n", c.title)
		catDir := filepath.Join(*outdir, c.title)
		err = os.MkdirAll(catDir, 0700)
		if err != nil {
			panic(err)
		}
		titleFile := filepath.Join(catDir, "00title.md")
		f, err := os.Create(titleFile)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(f, "# %s\n", c.title)
		f.Close()
		for _, r := range c.recipes {
			md, ok := markDown[r.link]
			if !ok {
				continue
			}
			mdFile := filepath.Join(catDir, r.title+".md")
			f, err := os.Create(mdFile)
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(f, "%s", md)
			f.Close()
			imgURL, ok := images[r.link]
			if !ok {
				continue
			}
			imgBase := filepath.Base(imgURL)
			imgFile := filepath.Join(*outdir, imgBase)
			_, err = os.Stat(imgFile)
			if err == nil {
				continue
			}
			f, err = os.Create(imgFile)
			if err != nil {
				panic(err)
			}
			resp, err := http.Get(imgURL)
			if err != nil {
				panic(err)
			}
			if resp.StatusCode/100 != 2 {
				fmt.Fprintf(os.Stderr, "%s: %s\n", imgURL, resp.Status)
				os.Exit(1)
			}
			_, err = io.Copy(f, resp.Body)
			if err != nil {
				panic(err)
			}
			resp.Body.Close()
			f.Close()
		}
	}
}

func recipe2md(n *html.Node) (img, md string) {
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			switch c.Data {
			case "h1":
				fmt.Fprintf(&sb, "## %s\n\n", nodeText(c))
			case "h2":
				fmt.Fprintf(&sb, "## %s\n\n", nodeText(c))
			case "h3":
				fmt.Fprintf(&sb, "### %s\n\n", nodeText(c))
			case "p":
				s := bufio.NewScanner(strings.NewReader(nodeText(c)))
				for s.Scan() {
					fmt.Fprintf(&sb, "%s\n", strings.TrimSpace(s.Text()))
				}
				fmt.Fprintf(&sb, "\n")
			case "ul":
				for _, li := range htmlelements.GetElementsByTagName(c, "li") {
					item := htmlelements.InnerText(li)
					item = strings.ReplaceAll(item, "\n", " ")
					item = compressSpace.ReplaceAllString(item, " ")
					item = strings.TrimSpace(item)
					fmt.Fprintf(&sb, "* %s\n", item)
				}
				fmt.Fprintf(&sb, "\n")
			case "span":
				if htmlelements.GetAttribute(c, "itemprop") == "image" {
					img = htmlelements.GetAttribute(c, "content")
					fmt.Fprintf(&sb, "![](%s)\n\n", path.Base(img))
				}
			default:
				if debug {
					fmt.Printf("%v\n", c)
				}
			}
		}
	}
	md = sb.String()
	return
}

func nodeText(n *html.Node) (ret string) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			ret += c.Data
		}
	}
	return
}
