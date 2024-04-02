# XMLLinksValidator

A tool that is written in Golang that concurrently fetches and validates the http links from the XML files (SiteMap currently). It can be further enhanced to validate any links in an xml file.

## Version

- Go 1.14.2

## Features

- Sitemap & Product Feed Validation Options
- Check for 200 response & errors
- Also, check for links which are inactive/inaccessible even though their status code is 200
- Generates the output CSV file with format: `URL, StatusCode, Validity, ErrorMessage`
- Customizable Concurrency level

## Project Structure

```
  xml-links-validator
    ├── ...
    └── dedupe                         # project for link validity test
        ├── dedupe.go                  # fetches all the unique links from the xml file and output them into txt file
    └── txtchecker
        ├── txtchecker.go              # validates all the unique links from feed.txt file and output the result into csv file
    └── Sitemap.xml                    # sample sitemap (replace this with your Sitemap XML)
```

### Running the project

Golang needs to be installed in order to run the project.

- Replace Sitemap.xml with the sitemap that needs to be validated
- Run dedupe.go command which fetches all the links into one text file
- Run txtchecker.go command which validates all the links & outputs csv file
- _Note: Concurrency level can be changed by updating workers value in txtchecker.go file._ _Default is 8_

```
    cd dedupe && go run dedupe.go
    cd ..
    cd txtchecker && go run txtchecker.go
```

For e.g.

```
    cd dedupe && go run dedupe.go -xml Sitemap

    cd ..
    cd txtchecker && go run txtchecker.go
```
