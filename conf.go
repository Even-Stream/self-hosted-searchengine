package main

import (
    "os"
    "encoding/json"
)

//Freq(frequency) is in days. Defaults to 60.
//Image controls if images should be indexed(WIP). Defaults to true.
type Seed struct {
    URL string         `json:"URL"`
    Freq int           `json:"Freq"`
    Images bool        `json:"Images"`
}

type Seed_cont struct {
    Seeds []Seed       `json:"Seeds"`
}

var Seed_URLs []Seed

func Load_seeds() []Seed {
    seed_file, err := os.ReadFile("seeds.json")
    Err_check(err)

    var cont Seed_cont
    json.Unmarshal(seed_file, &cont)

    return cont.Seeds
}