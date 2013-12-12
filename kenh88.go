package main

import (
	"github.com/PuerkitoBio/goquery"
)

func newKenh88() (si *kenh88) {
	si = new(kenh88)
	si.site = new(site)
	si.hostName = "www.kenh88.com"
	si._indexPath = ""
	si._title = "Desktop"
	return
}

type kenh88 struct {
	*site
}

func (si *kenh88) showsList(doc *goquery.Document, shs *shows) {
	doc.Find("#makers").Each(func(i int, s *goquery.Selection) {

		sh := new(show)

		sh.Id = i + 1
		sh.Title = s.Find("strong").Text()
		sh.Img, _ = s.Find("img").Attr("src")
		sh.Img = si.path(sh.Img)

		sh.Link, _ = s.Find("a").Attr("href")
		sh.Link = "/episodes" +
			"?show_p_url=" + si.pathEsc(sh.Link) +
			"&referer=" + si.pathEsc(sh.Link)

		shs.Shows = append(shs.Shows, sh)
	})
	shs.Title = si.title()
}
