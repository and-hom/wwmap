package main

import (
	"github.com/kokardy/saxlike"
	"os"
)

func main() {
	tagKey := os.Args[1]

	r := os.Stdin
	handler := &NodeSearchHandler{
		TagName:tagKey,
		WriteProgress:true,
		Node:false,
		Found:false,
	}
	parser := saxlike.NewParser(r, handler)
	parser.SetHTMLMode()
	parser.Parse()
}
