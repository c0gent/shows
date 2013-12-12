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

type contextInfo struct {
	hostName string
	path     string
}

var Ke *kenh88
var Ph *phimmobile

var net *unframed.NetHandle

func main() {

	Ph = newPhimmobile()

	Ke = newKenh88()

	net = unframed.NewNet()

	unframed.DefaultPageTitle = "Rose's Shows"
	net.TemplateFiles(
		"tmpl/base.html.tmpl",
		"tmpl/show-list.html.tmpl",
		"tmpl/episode-list.html.tmpl",
		"tmpl/_show-list_table.html.tmpl",
	)
	net.Get("/episodes/watch", episodesWatch)
	net.Get("/episodes", episodesList)
	net.Dir("assets/")
	net.Dir("public/")
	net.Get("/", showsList)

	net.LoadTemplates()

	log.Message("Serving Rose's Shows")
	net.Serve("80")

}

type site struct {
	hostName   string
	_indexPath string
	_title     string
}

func (si *site) title() string {
	return si.hostName + " - " + si._title
}
func (si *site) rootPath(p ...string) string {
	return "http://" + si.hostName + strings.Join(p, "")
}
func (si *site) path(p ...string) string {
	return si.rootPath(si._indexPath + strings.Join(p, ""))
}
func (si *site) pathEsc(p ...string) string {
	return url.QueryEscape(si.path(p...))
}
func (si *site) rootPathEsc(p ...string) string {
	return url.QueryEscape(si.rootPath(p...))
}
func (si *site) pathUnesc(p string) string {
	s, err := url.QueryUnescape(p)
	if err != nil {
		log.Error(err)
		return "invalid path unescapage and stuff"
	}
	return s
}
func (si *site) esc(s string) string {
	return url.QueryEscape(s)
}
func (si *site) createShowLink() {

}

func (si *site) createEpisodeLink() {

}

type ScrapeSite interface {
	title() string
	path(...string) string
	pathEsc(...string) string
	showsList(*goquery.Document, *shows)
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

	net.ExeTmpl(w, "showList",
		scrapeShows(Ph),
		//scrapeShows(Ke),
	)
}

func scrapeShows(si ScrapeSite) (shs shows) {
	var doc *goquery.Document
	var e error

	if doc, e = goquery.NewDocument(si.path("")); e != nil {
		panic(e.Error())
	}

	si.showsList(doc, &shs)
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

	query := r.URL.Query()
	spu := query.Get("show_p_url")
	imgUrl := query.Get("img_url")
	showTitle := query.Get("show_title")

	//log.Message("spu:", spu)
	fakeShowLink := spu //restoreURL

	//log.Message("fakeShowLink:", fakeShowLink)
	showUrl := actualShowLink(fakeShowLink)

	psu, _ := url.Parse(showUrl)
	//log.Message("episodesList() showUrl:", showUrl, "Hostname:", psu.Host)

	var doc *goquery.Document
	var err error
	eps := make(episodes, 0)

	fakeMobile = true
	if fakeMobile {
		doc, err = NewMobileDocument(showUrl, showUrl)
	} else {
		doc, err = goquery.NewDocument(showUrl)
	}

	if err != nil {
		panic(err)
	}

	episodeContainer := doc.Find("#countrydivcontainer")

	episodeContainer.Find("div").Each(func(i int, s *goquery.Selection) {
		ep := new(episode)
		ep.Id = i + 1
		ep.Title = s.Find("strong").Text()
		ep.Link, _ = s.Find("a").Attr("href")
		ep.Link = "episodes/watch" +
			"?episode_path=" + url.QueryEscape(ep.Link) +
			"&hostname=" + url.QueryEscape(psu.Host)

		eps = append(eps, ep)

	})

	net.ExeTmpl(w, "episodeList", eps, imgUrl, showTitle)

}

func actualShowLink(fakeShowLink string) (asl string) {
	var doc *goquery.Document
	var e error
	//log.Message("actualShowLink(): fakeShowLink:", fakeShowLink)
	if doc, e = goquery.NewDocument(fakeShowLink); e != nil {
		panic(e)
	}
	psu, _ := url.Parse(fakeShowLink)
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

func scrapeEpisodesList() {
	// unused right now
}

func episodesWatch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	ep := query.Get("episode_path")
	ho := query.Get("hostname")

	vidUrl := scrapeVidUrl(ho, ep)

	http.Redirect(w, r, vidUrl, http.StatusFound)
}

func scrapeVidUrl(hostname string, epath string) (avl string) {
	//var fakeMobile bool = false

	episodeUrl := "http://" + hostname + epath
	//log.Message("scrapeVidUrl() episodeUrl:", episodeUrl)

	//var doc *goquery.Document
	//var err error
	doc, err := NewMobileDocument(episodeUrl, episodeUrl)
	if err != nil {
		log.Message(err)
	}

	var iFrameUrl string

	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		//log.Message("Contains:", strings.Contains(src, "/typo.php?"), "-- full:", src)

		if strings.Contains(src, "/typo.php?") {
			iFrameUrl = src
			//log.Message("iFrameUrl:", src)
		}
	})

	doc, err = NewMobileDocument("http://"+hostname+iFrameUrl, episodeUrl)
	if err != nil {
		log.Message(err)
	}

	//log.Message(list1Url+iFrameUrl, ":", doc.Text()) // *** Lots of good info here ***

	doc.Find("link").Each(func(i int, s *goquery.Selection) {

		var href string
		href, _ = s.Attr("href")
		if strings.Contains(href, "dailymotion.com") {
			avl = href
			log.Message("scrapeVidUrl(): Actual Video Link:", avl)
		}
		return
	})

	return
}

func NewMobileDocument(u string, referer string) (d *goquery.Document, err error) {
	//log.Message("calling newMobilDocument():", u)
	ua := "Mozilla/5.0 (Linux; U; Android 4.0.2; en-us; Galaxy Nexus Build/ICL53F) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30 DNT: 1"

	req, err := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Referer", referer)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	d, err = goquery.NewDocumentFromResponse(res)
	return
}

/*

var urlSlash string = "/"
var urlSlashRepl string = "_~s~_"
var urlQmark string = "?"
var urlQmarkRepl string = "_~q~_"
var urlAnpr string = "&"
var urlAnprRepl string = "_~a~_"
var urlColon string = ":"
var urlColonRepl string = "_~c~_"
var urlEqual string = ":"
var urlEqualRepl string = "_~e~_"

func passableUrl(u string) (pass string) {
	//log.Message(url)
	pa := strings.Split(u, urlSlash)
	pass = strings.Join(pa, urlSlashRepl)

	pa = strings.Split(pass, urlQmark)
	pass = strings.Join(pa, urlQmarkRepl)

	pa = strings.Split(pass, urlAnpr)
	pass = strings.Join(pa, urlAnprRepl)

	pa = strings.Split(pass, urlColon)
	pass = strings.Join(pa, urlColonRepl)

	pa = strings.Split(pass, urlEqual)
	pass = strings.Join(pa, urlEqualRepl)
	//log.Message(pass)

	pass = url.QueryEscape(pass)
	return
}

func restoreUrl(pass string) (u string) {
	//log.Message("restoreUrl msg1", pass)
	pa := strings.Split(pass, urlSlashRepl)
	u = strings.Join(pa, urlSlash)

	pa = strings.Split(u, urlQmarkRepl)
	u = strings.Join(pa, urlQmark)

	pa = strings.Split(u, urlAnprRepl)
	u = strings.Join(pa, urlAnpr)

	pa = strings.Split(u, urlColonRepl)
	u = strings.Join(pa, urlColon)

	pa = strings.Split(u, urlEqualRepl)
	u = strings.Join(pa, urlEqual)

	u, _ = url.QueryUnescape(u)
	//log.Message("restoreUrl msg2", url)
	return
}

func validUrlPath(path string) (validPath string) {
	p := []rune(path)
	if p[0] != []rune("/")[0] {
		validPath = "/" + string(p)
	} else {
		validPath = string(p)
	}
	return
}


*/
