package main

import (
	//"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/nsan1129/unframed"
	"github.com/nsan1129/unframed/log"
	"net/http"
	"net/url"
	"strings"
	//"strconv"
)

var list1Url string = "http://www.phimmobile.com"

var net *unframed.NetHandle

func main() {

	net = unframed.NewNet()

	unframed.DefaultPageTitle = "Rose's Shows"
	net.TemplateFiles(
		"tmpl/base.html.tmpl",
		"tmpl/show-list.html.tmpl",
		"tmpl/episode-list.html.tmpl",
		"tmpl/_show-list_table.html.tmpl",
	)
	net.Get("/episodes/{showLink}", episodesList)
	net.Get("/episodes/watch/{episodeUrl}", episodeWatch)
	net.Dir("assets/")
	net.Dir("public/")
	net.Get("/", showsList)

	net.LoadTemplates()

	log.Message("Serving Rose's Shows")
	net.Serve("8080")

}

type shows struct {
	Shows []*show
	Title string
}
type show struct {
	Id    int
	Img   string
	Title string
	Link  string
}

func showsList(w http.ResponseWriter, r *http.Request) {

	//list2Url := "http://www.kenh88.com"
	net.ExeTmpl(w, "showList", (scrapeShows(list1Url, list1Url+"/index.php", "Mobile - ")))
}

func scrapeShows(rootUrl string, indexUrl string, titlePre string) (shs shows) {
	var doc *goquery.Document
	var e error

	if doc, e = goquery.NewDocument(indexUrl); e != nil {
		panic(e.Error())
	}

	doc.Find("#makers").Each(func(i int, s *goquery.Selection) {
		sh := new(show)

		sh.Id = i + 1
		sh.Title = s.Find("strong").Text()
		sh.Img, _ = s.Find("img").Attr("src")
		sh.Img = rootUrl + sh.Img

		sh.Link, _ = s.Find("a").Attr("href")
		sh.Link = "/episodes/" + passableUrl(rootUrl+sh.Link)

		shs.Shows = append(shs.Shows, sh)
	})
	shs.Title = strings.Split(rootUrl, urlSlash)[2]
	return
}

func actualShowLink(showLink string) (asl string) {
	var doc *goquery.Document
	var e error

	if doc, e = goquery.NewDocument(showLink); e != nil {
		panic(e)
	}
	psu, _ := url.Parse(showLink)
	baseUrl := psu.Scheme + "://" + psu.Host

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		srcAttr, _ := s.Attr("src")
		if srcAttr == "image/watch_now_new.png" || srcAttr == "image/watch_new.png" {
			var docA *goquery.Document
			var e error
			actualShowUrl, _ := s.Parent().Attr("href")

			if docA, e = goquery.NewDocument(baseUrl + actualShowUrl); e != nil {
				panic(e.Error())
			}
			_ = docA

			//ep := new(episode)
			asl = baseUrl + actualShowUrl

			//net.ExeTmpl(w, "episodeList", eps)
		}
	})
	return
}

type episodes []*episode
type episode struct {
	Id    int
	Title string
	Link  string
}

func episodesList(w http.ResponseWriter, r *http.Request) {
	var fakeMobile bool = false
	//log.Message("episodeList(): r.URL:", r.URL)
	//log.Message("episodeList(): net.UrlVar:", net.UrlVar("showLink", r))

	//log.Message("episodeList msg1", restoreUrl(net.UrlVar("showLink", r)))

	fakeShowLink := restoreUrl(net.UrlVar("showLink", r))

	showUrl := actualShowLink(fakeShowLink)
	psu, _ := url.Parse(showUrl)

	log.Message("episodeList() showUrl:", showUrl, "Hostname:", psu.Host)

	var doc *goquery.Document
	var err error
	eps := make(episodes, 0)

	fakeMobile = true
	if fakeMobile {
		doc, err = NewMobileDocument(showUrl, fakeShowLink)
	} else {
		doc, err = goquery.NewDocument(showUrl)
	}

	if err != nil {
		panic(err)
	}

	episodeContainer := doc.Find("#countrydivcontainer")

	baseUrl := psu.Scheme + "://" + psu.Host
	episodeContainer.Find("div").Each(func(i int, s *goquery.Selection) {
		ep := new(episode)
		ep.Id = i + 1
		ep.Title = s.Find("strong").Text()
		ep.Link, _ = s.Find("a").Attr("href")
		ep.Link = "watch/" + passableUrl(baseUrl+ep.Link)
		eps = append(eps, ep)

	})

	net.ExeTmpl(w, "episodeList", eps)

}

func scrapeEpisodesList() {
	// unused right now
}

func episodeWatch(w http.ResponseWriter, r *http.Request) {
	vidUrl := scrapeVidUrl(restoreUrl(net.UrlVar("episodeUrl", r)))
	http.Redirect(w, r, vidUrl, http.StatusFound)
}

func scrapeVidUrl(episodeUrl string) (avl string) {
	//var fakeMobile bool = false
	log.Message("scrapeVidUrl() episodeLink:", episodeUrl)

	//var doc *goquery.Document
	//var err error
	doc, err := NewMobileDocument(episodeUrl, episodeUrl)
	if err != nil {
		log.Message(err)
	}

	/*
		fakeMobile = true
		if fakeMobile {
			doc, err = NewMobileDocument(episodeUrl)
		} else {
			doc, err = goquery.NewDocument(episodeUrl)
		}
	*/
	iFrameUrl := ""

	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		//log.Message("Contains:", strings.Contains(src, "/typo.php?"), "-- full:", src)

		if strings.Contains(src, "/typo.php?") {
			iFrameUrl = src
			log.Message("iFrameUrl:", src)
		}
	})

	doc, err = NewMobileDocument(list1Url+iFrameUrl, episodeUrl)
	if err != nil {
		log.Message(err)
	}

	//log.Message(list1Url+iFrameUrl, ":", doc.Text()) // *** Lots of good info here ***

	doc.Find("link").Each(func(i int, s *goquery.Selection) {

		var href string
		href, _ = s.Attr("href")
		if strings.Contains(href, "dailymotion.com") {
			log.Message("avl:", avl)
			avl = href
		}
		return
	})

	return
}

var urlSlash string = "/"
var urlSlashRepl string = "__"
var urlQmark string = "?"
var urlQmarkRepl string = "_~_"

func passableUrl(url string) (pass string) {
	//log.Message(url)
	pa := strings.Split(url, urlSlash)
	pass = strings.Join(pa, urlSlashRepl)

	pa = strings.Split(pass, urlQmark)
	pass = strings.Join(pa, urlQmarkRepl)
	//log.Message(pass)
	return
}

func restoreUrl(pass string) (url string) {
	//log.Message("restoreUrl msg1", pass)
	pa := strings.Split(pass, urlSlashRepl)
	url = strings.Join(pa, urlSlash)

	pa = strings.Split(url, urlQmarkRepl)
	url = strings.Join(pa, urlQmark)
	//log.Message("restoreUrl msg2", url)
	return
}

func NewMobileDocument(url string, referer string) (d *goquery.Document, err error) {
	//log.Message("calling newMobilDocument")
	ua := "Mozilla/5.0 (Linux; U; Android 4.0.2; en-us; Galaxy Nexus Build/ICL53F) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30 DNT: 1"

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Referer", referer)
	res, err := http.DefaultClient.Do(req)
	defer res.Body.Close()
	d, err = goquery.NewDocumentFromResponse(res)
	return
}
