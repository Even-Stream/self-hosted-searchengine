package main

import (
    "fmt"
    "log"
    "time"
    "sync"
    "os"

    "github.com/blevesearch/bleve/v2"
    "github.com/blevesearch/bleve/v2/registry"
)

func Err_check(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func Setup_crawl(now string) {
    var wg sync.WaitGroup
    
    seeds := Load_seeds()
    seed_chan := make(chan Seed, len(seeds))
    
    crawl_num := 3
    for w := 0; w < 3; w++ {
        wg.Add(1)
        go func() {defer wg.Done(); Queue_seed(seed_chan, crawl_num, now)}()
    }
    for _, seed := range seeds {seed_chan <- seed}
    close(seed_chan)
    wg.Wait()
 
    fmt.Println("Seeds queued.")
}

func DB_create(path string) {
    if _, err := os.Stat(path); err != nil {
        file, err := os.Create(path)
        Err_check(err)
        file.Close()
    }
}

func init() {
    registry.RegisterTokenizer(TokenName, tokenizerConstructor)
    now := time.Now().UTC().Format(time.DateOnly)

    var err error
    DB_create("./external_crawl.db")

    opt := Option{
        Index: "index.bleve",
        Opt:   "search-hmm",
        Trim:  "trim",
    }

    Index, err = bleve.Open("index.bleve")
    if err == bleve.ErrorIndexPathDoesNotExist {
        Index, err = BuildNewIndex(opt)
        Err_check(err)
    } else {Err_check(err)}

    Setup_crawl(now)

    for w := 0; w < 3; w++ {
        go Initiate_crawl(w, now)
    }
}

func main() {
    Listen()
}