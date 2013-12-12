package main

import (
	"github.com/PuerkitoBio/goquery"
)

func newPhimmobile() *phimmobile {
	si := new(phimmobile)
	si.site = new(site)
	si.hostName = "www.phimmobile.com"
	si._indexPath = "/index.php"
	si._title = "Mobile"
	return si
}

type phimmobile struct {
	*site
}

func (si *phimmobile) showsList(doc *goquery.Document, shs *shows) {
	doc.Find("#makers").Each(func(i int, s *goquery.Selection) {

		sh := new(show)

		sh.Id = i + 1
		sh.Title = s.Find("strong").Text()
		sh.Img, _ = s.Find("img").Attr("src")
		sh.Img = si.rootPath(sh.Img)

		sh.Link, _ = s.Find("a").Attr("href")
		sh.Link = "/episodes" +
			"?show_p_url=" + si.pathEsc(sh.Link) +
			"&referer=" + si.pathEsc(sh.Link) +
			"&img_url=" + sh.Img +
			"&show_title=" + si.esc(sh.Title)

		shs.Shows = append(shs.Shows, sh)
	})
	shs.Title = si.title()
}
