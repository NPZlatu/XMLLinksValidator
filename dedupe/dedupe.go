package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

func main() {
	xmlFile := "../Sitemap.xml"
	feedFile := "../feed.txt"
	parentXmlNode := "url"

	inFile, err := os.Open(xmlFile)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	outFile, err := os.Create(feedFile)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// trick for deduping/uniqueness in Go: create a map where the key is the
	// item to be deduped and the value is a bool and use that to store only
	// unique values as the keys
	uniqueLinks := make(map[string]bool)

	d := xml.NewDecoder(inFile)
	for {
		t, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		switch t := t.(type) {
		case xml.StartElement:
			if t.Name.Local != parentXmlNode {
				continue
			}
			// Extract URL from Sitemap
			var xmlItemKey string
			for _, attr := range t.Attr {
				if attr.Name.Local == "loc" {
					xmlItemKey = strings.TrimSpace(attr.Value)
					break
				}
			}
			uniqueLinks[xmlItemKey] = true
		}
	}

	// write unique URLs to new feed file
	for url := range uniqueLinks {
		fmt.Println("writing URL", url)
		feedLine := fmt.Sprintf("%s\n", url)
		if _, err := outFile.Write([]byte(feedLine)); err != nil {
			panic(err)
		}
	}
}
