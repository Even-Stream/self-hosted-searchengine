package main

import (
    "time"
    "net/http"
)

func Listen() {
    mux := http.NewServeMux()
    mux.Handle("/", http.FileServer(http.Dir("./frontend")))
    mux.HandleFunc("/search/", Search)
    mux.HandleFunc("/cache/", Cache_get)

    srv := &http.Server {
        Addr: ":1025",
        IdleTimeout: 10 * time.Second,
        ReadTimeout: 6 * time.Second,
        ReadHeaderTimeout: 6 * time.Second,
        WriteTimeout: 10 * time.Second,
        Handler: mux,
    }
    srv.ListenAndServe()
}
