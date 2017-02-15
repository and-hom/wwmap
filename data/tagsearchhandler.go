package main

import (
	"encoding/xml"
	"fmt"
	"bytes"
	"github.com/kokardy/saxlike"
	"os"
	"strings"
)


//VoidHandler is a implemented Handler that do nothing.
type NodeSearchHandler struct {
	saxlike.VoidHandler
	TagName       string
	WriteProgress bool

	cnt           int
	Node          bool
	Found         bool
	Buffer        bytes.Buffer
}

func (h *NodeSearchHandler) StartElement(element xml.StartElement) {
	if element.Name.Local == "node" {
		h.cnt += 1
		h.Node = true
	}
	if h.Node {
		h.Buffer.WriteString("<" + element.Name.Local)
		for _, attr := range element.Attr {
			h.Buffer.WriteString(" " + attr.Name.Local + "=\"" + strings.Replace(attr.Value,"\"","&quot;",-1) + "\"")
		}
		h.Buffer.WriteString(">")
	}
	if h.Node && element.Name.Local == "tag" && hasEqAttr(element.Attr, "k", h.TagName) {
		h.Found = true
	}
}

func hasEqAttr(attrs []xml.Attr, name string, value string) bool {
	for _, attr := range attrs {
		if attr.Name.Local == name && attr.Value == value {
			return true
		}
	}
	return false
}

func (h *NodeSearchHandler) EndElement(element xml.EndElement) {
	if h.Node {
		h.Buffer.WriteString("</" + element.Name.Local + ">")
	}
	if element.Name.Local == "node" {
		if h.WriteProgress && h.cnt % 100000 == 0 {
			fmt.Fprintf(os.Stderr, "%d\n", h.cnt)
		}
		if h.Found {
			fmt.Println(h.Buffer.String())
		}
		h.Node = false
		h.Found = false
		h.Buffer.Reset()
	}
}

func (h *NodeSearchHandler) CharData(char xml.CharData) {
	if h.Node {
		h.Buffer.Write([]byte(char))
	}
}
