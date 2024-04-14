package main

import (
    //"fmt"
    "html"
    "strings"
    "regexp"
    "text/template"
    "net/http"

    "github.com/blevesearch/bleve/v2"
)

type Result struct {
    Title, URL, Highlight string
}

var htmlreg1 = regexp.MustCompile(`(&lt;|\A).*?&gt;`)
var htmlreg2 = regexp.MustCompile(`&lt;.*?(&gt;|\z)`)

func stripHtmlRegex(s string) string {
    r := htmlreg1.ReplaceAllString(s, "")
    r = htmlreg2.ReplaceAllString(r, "")
    r = strings.NewReplacer("&gt;", "", "&lt;", "").Replace(r)
    return html.UnescapeString(r)
}

func Search(w http.ResponseWriter, req *http.Request) {
    query_string := req.FormValue("query")

    query := bleve.NewMatchQuery(query_string)

    searchRequest := bleve.NewSearchRequest(query)
    searchRequest.Fields = []string{"Mime", "Title"}
    searchRequest.Highlight = bleve.NewHighlight()
    searchRequest.Size = 20

    searchResult, err := Index.Search(searchRequest)
    Err_check(err)

    var results []Result    
    for _, hit := range searchResult.Hits {
        //fmt.Println(hit.Expl.String())
        cres := Result{URL: hit.ID}
        if title := hit.Fields["Title"]; len(title.(string)) != 0 {
            cres.Title = title.(string)
        } else {cres.Title = hit.ID}

        if frags := hit.Fragments["Body"]; len(frags) > 0 {cres.Highlight = stripHtmlRegex(frags[0])}
        
        results = append(results, cres)
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    build_respage(w, results)
}


func Cache_get(w http.ResponseWriter, req *http.Request) {
    doc_id := req.FormValue("id")

    query := bleve.NewDocIDQuery([]string{doc_id})
    searchRequest := bleve.NewSearchRequest(query)
    searchRequest.Fields = []string{"Body"}

    searchResult, err := Index.Search(searchRequest)
    Err_check(err)

    if len(searchResult.Hits) == 0 {return}

    cache := searchResult.Hits[0].Fields["Body"].(string)

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(cache))
}

func build_respage(w http.ResponseWriter, results []Result) {
    resulttemp := template.New("results.html")
    resulttemp, err := resulttemp.ParseFiles("./templates/results.html")
    Err_check(err)

    resulttemp.Execute(w, results)
}
