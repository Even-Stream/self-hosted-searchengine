package main

import (
    "github.com/blevesearch/bleve/v2"
    "github.com/blevesearch/bleve/v2/mapping"
    "github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
    "github.com/blevesearch/bleve/v2/analysis/char/html"
    "github.com/blevesearch/bleve/v2/analysis/token/lowercase"
)

var Index bleve.Index

type Record struct {
    Type, Mime, Host, Title, Body, LastCrawl, NextCrawl string
}

type Option struct {
	Index                  string
	Dicts, Opt, Trim string
	Alpha                  bool
	Name, Sep              string
}

func NewMapping(opt Option) (*mapping.IndexMappingImpl, error) {
	mapping := bleve.NewIndexMapping()

	bodyFieldMapping := bleve.NewTextFieldMapping()
	bodyFieldMapping.Analyzer = TokenName

	pageMapping := bleve.NewDocumentMapping()
        pageMapping.AddFieldMappingsAt("Body", bodyFieldMapping)

        mapping.AddDocumentMapping("page", pageMapping)

        err := mapping.AddCustomTokenizer(TokenName, map[string]interface{}{
            "type":  TokenName,
            "dicts": opt.Dicts,
            "opt":   opt.Opt,            
            "trim":  opt.Trim,
            "alpha": opt.Alpha,
        })
	Err_check(err)

        err = mapping.AddCustomAnalyzer(TokenName, map[string]interface{}{
            "type": custom.Name,
            "char_filters": []string{
                html.Name,
            },
            "tokenizer": TokenName,
            "token_filters": []string{
                lowercase.Name,
            },
        })
	Err_check(err)

        mapping.TypeField = "Type"
	mapping.DefaultAnalyzer = "standard"
	return mapping, nil
}

func BuildNewIndex(opt Option) (bleve.Index, error) {
	var (
		mapping *mapping.IndexMappingImpl
		err      error
	)

	mapping, err = NewMapping(opt)
	Err_check(err)

	return bleve.New(opt.Index, mapping)
}