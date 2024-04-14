package main

import (
    "fmt"
    //"log"
    "time"
    "strings"
    "strconv"
    "regexp"
    "math/rand/v2"
    "net/http"
    "net/url"
    "crypto/tls"

    "github.com/gabriel-vasile/mimetype"
    "github.com/gocolly/colly/v2"
    "github.com/gocolly/colly/v2/extensions"
    "github.com/gocolly/colly/v2/queue"
    "github.com/velebak/colly-sqlite3-storage/colly/sqlite3"
)

var charset_replace = regexp.MustCompile(`charset=(.*)\>`)
var spam = regexp.MustCompile(`watch\.impress\.co\.jp`)
var google_accounts = regexp.MustCompile(`accounts\.google\.com`)
var file_types = regexp.MustCompile(`\.(png|jpg|jpeg|gif|ico|pdf|iso|exe|msi)$`)

func takingbreak() {

}

func AddURLWD(q *queue.Queue, URL string, depth int) error {
	u, err := url.Parse(URL)
	if err != nil {
		return err
	}

        //this might be where max depth could be added
	r := &colly.Request{
                URL: u,
		Method: "GET",
                Depth: depth,
                Ctx: colly.NewContext(),
	}
        r.Ctx.Put("Freq", "60")
        r.Ctx.Put("Images", "true")

	return q.AddRequest(r)
}

func Crawl(db_path, crawl_time string, worker int) {
    md := 3

    title_func := func(e *colly.HTMLElement) {
        e.Request.Ctx.Put("title", e.Text)
    }

    request_func := func(r *colly.Request) {
        doc, err := Index.Document(r.URL.String())
        Err_check(err)
        if doc != nil {
            //fmt.Println("Already crawled: ", r.URL.String())
            r.Abort()
        } //else {fmt.Println("Depth: ", r.Depth , "Visiting: ", r.URL)}
    }

    response_func := func(r *colly.Response) {
        //fmt.Println("Response from: ", r.Request.URL.String())
        rec_type := mimetype.Detect(r.Body).String()

        result_type := "other" 
        if strings.HasPrefix(rec_type, "text") {result_type = "page"}
   
        corr_body := charset_replace.ReplaceAll(r.Body, []byte(`charset=UTF-8">`))

        r.Ctx.Put("type", result_type)
        r.Ctx.Put("mime", rec_type)
        r.Ctx.Put("host", r.Request.Host)
        r.Ctx.Put("body", string(corr_body))
        r.Ctx.Put("lastcrawl", crawl_time)
    }

    err_func := func(_ *colly.Response, err error) {
        //log.Println("Something went wrong:", err)
    }

    scraped_func := func(r *colly.Response) {
        if r.Ctx.Get("type") != "page" {return}

        lct, err := time.Parse(time.DateOnly, r.Ctx.Get("lastcrawl"))
        Err_check(err)
        freq, err := strconv.Atoi(r.Request.Ctx.Get("Freq"))
        Err_check(err)

        data := Record{Type: r.Ctx.Get("type"), Mime: r.Ctx.Get("mime"), 
            Host: r.Ctx.Get("host"), Title: r.Ctx.Get("title"), Body: r.Ctx.Get("body"), 
            LastCrawl: r.Ctx.Get("lastcrawl"), NextCrawl: lct.Add(time.Hour * time.Duration(24 * freq)).Format(time.DateOnly)}

        go Index.Index(r.Request.URL.String(), data)
    }


    storage := &sqlite3.Storage{
        Filename: db_path,
    }
    ex_storage := &sqlite3.Storage{
        Filename: "./external_crawl.db",
    }

    q, _ := queue.New(1, storage)
    ex_q, _ := queue.New(1, ex_storage)

    c := colly.NewCollector(
        colly.DetectCharset(),
        colly.MaxDepth(md),
        colly.IgnoreRobotsTxt(),
        colly.DisallowedURLFilters(spam, google_accounts, file_types),
    )

    extensions.RandomUserAgent(c)
    c.SetRequestTimeout(time.Second * 10)

    c.WithTransport(&http.Transport{
        DisableKeepAlives: true,
    })

    c.MaxBodySize = 1024 * 1024
    c.AllowURLRevisit = false
    c.DisableCookies()

    c.Limit(&colly.LimitRule{
        DomainGlob: "*",
        Delay:    10 * time.Second,
        RandomDelay: 20 * time.Second,
    })

    c.WithTransport(&http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    })

    //callbacks

    c.OnHTML("a[href]", func(e *colly.HTMLElement) {
        link := strings.Split(e.Request.AbsoluteURL(e.Attr("href")), "#")[0]
        next_depth := e.Request.Depth + 1
   
        cq := q
        c_host := e.Request.Host

        match, _ := regexp.MatchString(`.*` + c_host + `.*`, link)
        if !match {cq = ex_q}

        match, _ = regexp.MatchString(link + `\#.*`, link)
        if !match {AddURLWD(cq, link, next_depth)}
    })

    c.OnRequest(request_func)
    c.OnResponse(response_func)

    c.OnHTML("title", title_func)

    c.OnError(err_func)
    c.OnScraped(scraped_func)


    ex_c := c.Clone()
    ex_c.OnRequest(request_func)
    ex_c.OnResponse(response_func)

    ex_c.OnHTML("title", title_func)

    ex_c.OnError(err_func)
    ex_c.OnScraped(scraped_func)
    
    q.Run(c)
    ex_q.Run(ex_c)

    //fmt.Printf("Crawer done: %d\n", worker)
}

func Initiate_crawl(worker int, now string) {
    db_path := fmt.Sprintf("./seed_crawl%d.db", worker)
    Crawl(db_path, now, worker)
}

func Queue_seed(seed_chan <-chan Seed, crawl_num int, crawl_time string) {

    for seed := range seed_chan {
        worker := rand.IntN(crawl_num)
        db_path := fmt.Sprintf("./seed_crawl%d.db", worker)
        DB_create(db_path)

        storage := &sqlite3.Storage{
            Filename: db_path,
        }
        q, _ := queue.New(1, storage)
        AddURLWD(q, seed.URL, 1)
    }
}
