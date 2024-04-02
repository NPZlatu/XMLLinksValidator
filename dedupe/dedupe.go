package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	xmlFile := "../sitemap.xml"
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
			for {
				t, err := d.Token()
				if err != nil {
					if err == io.EOF {
						break
					}
					panic(err)
				}
				switch t := t.(type) {
				case xml.EndElement:
					if t.Name.Local == "url" {
						if xmlItemKey != "" {
							uniqueLinks[xmlItemKey] = true
						}
						break
					}
				case xml.StartElement:
					if t.Name.Local == "loc" {
						// Extract URL
						innerToken, err := d.Token()
						if err != nil {
							panic(err)
						}
						if charData, ok := innerToken.(xml.CharData); ok {
							xmlItemKey = strings.TrimSpace(string(charData))
						}
					}
				}
			}
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
