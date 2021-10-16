package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/pkg/errors"
	"os"
	"strconv"
	"strings"
)

type property struct {
	Category              string
	Name                  string
	Description           string
	Type                  string
	Default               string
	ValidValues           string
	Importance            string
	DynamicUpdateMode     string
	ServerDefaultProperty string
}

func main() {

	kafkaVersion := map[string]string{
		"0101": "0.10.1",
		"0102": "0.10.2",
		"0110": "0.11.0",
		"10":   "1.0",
		"11":   "1.1",
		"20":   "2.0",
		"21":   "2.1",
		"22":   "2.2",
		"23":   "2.3",
		"24":   "2.4",
		"25":   "2.5",
		"26":   "2.6",
		"27":   "2.7",
		"28":   "2.8",
		"30":   "3.0",
	}

	for k, versionLabel := range kafkaVersion {
		var properties []property
		docVersion, _ := strconv.Atoi(k)
		if docVersion < 24 || docVersion > 100 {
			properties = getDocBefore24(k)
		} else if docVersion < 25 {
			properties = getDocBefore25(k)
		} else {
			properties = getDoc(k)
		}

		bytes, err := json.Marshal(properties)
		if err != nil {
			panic(err)
		}
		f, err := os.Create("kafka-doc-" + versionLabel + ".json")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		_, err = w.Write(bytes)
		if err != nil {
			panic(err)
		}
	}
}

func getDoc(versionCode string) []property {
	var properties []property
	c := colly.NewCollector()

	c.OnHTML(`#configuration-template`, func(e *colly.HTMLElement) {
		document, err := goquery.NewDocumentFromReader(strings.NewReader(e.Text))
		if err != nil {
			panic(err)
		}

		var currentCategory string

		document.Find(".config-list > li, h3 > a").Each(func(i int, selection *goquery.Selection) {
			p := property{}

			if selection.Nodes[0].Data == "a" {
				categoryLinkAttr, _ := selection.Attr("id")
				if categoryLinkAttr == "" {
					return
				}
				currentCategory = strings.Replace(categoryLinkAttr, "configs", "", 1)
				return
			}

			p.Category = currentCategory

			// property name is the last <a> for v2.5
			propKey := selection.Find("h4 > a").Last()
			p.Name = propKey.Text()

			html, _ := selection.Html()

			description := strings.Split(strings.Split(html, "</h4>")[1], "<table")[0]
			p.Description = description

			selection.Find("table > tbody").Each(func(i int, selection *goquery.Selection) {
				selection.Find("tr").Each(func(i int, selection *goquery.Selection) {
					prop := selection.Find("th").First().Text()
					value := selection.Find("td").First().Text()

					// we remove the ":" at the end of the string
					prop = prop[:len(prop)-1]

					switch prop {
					case "Type":
						p.Type = value
					case "Default":
						p.Default = value
					case "Valid Values":
						p.ValidValues = value
					case "Importance":
						p.Importance = value
					case "Update Mode":
						p.DynamicUpdateMode = value
					case "Server Default Property":
						p.ServerDefaultProperty = value
					}
				})
			})
			properties = append(properties, p)
		})
	})
	c.OnError(func(response *colly.Response, e error) {
		if response.StatusCode == 404 {
			panic(errors.Errorf("Not found"))
		}
		panic(errors.Wrapf(e, "colly error (status:%d)", response.StatusCode))
	})
	err := c.Visit("https://kafka.apache.org/" + versionCode + "/documentation.html")
	if err != nil {
		panic(err)
	}

	return properties
}

func getDocBefore25(versionCode string) []property {
	var properties []property
	c := colly.NewCollector()

	c.OnHTML(`#configuration-template`, func(e *colly.HTMLElement) {
		document, err := goquery.NewDocumentFromReader(strings.NewReader(e.Text))
		if err != nil {
			panic(err)
		}

		var currentCategory string

		document.Find(".config-list > li, h3 > a").Each(func(i int, selection *goquery.Selection) {
			p := property{}

			if selection.Nodes[0].Data == "a" {
				categoryLinkAttr, _ := selection.Attr("id")
				currentCategory = strings.Replace(categoryLinkAttr, "configs", "", 1)
				return
			}

			p.Category = currentCategory

			propKey := selection.Find("b").First()
			p.Name = propKey.Text()
			html, _ := selection.Html()

			description := strings.Split(strings.Split(html, "</b>: ")[1], "<br/>")[0]
			p.Description = description

			selection.Find(".horizontal-list").Each(func(i int, selection *goquery.Selection) {
				selection.Find("li").Each(func(i int, selection *goquery.Selection) {
					html, _ := selection.Html()

					prop := selection.Find("b").First().Text()
					value := strings.Split(html, "</b>: ")[1]

					switch prop {
					case "Type":
						p.Type = value
					case "Default":
						p.Default = value
					case "Valid Values":
						p.ValidValues = value
					case "Importance":
						p.Importance = value
					case "Update Mode":
						p.DynamicUpdateMode = value
					case "Server Default Property":
						p.ServerDefaultProperty = value
					}
				})
			})
			properties = append(properties, p)
		})
	})
	c.OnError(func(response *colly.Response, e error) {
		if response.StatusCode == 404 {
			panic(errors.Errorf("Not found"))
		}
		panic(errors.Wrapf(e, "colly error (status:%d)", response.StatusCode))
	})
	err := c.Visit("https://kafka.apache.org/" + versionCode + "/documentation.html")
	if err != nil {
		panic(err)
	}

	return properties
}

func getDocBefore24(versionCode string) []property {
	var properties []property
	c := colly.NewCollector()
	c.OnHTML(`#configuration-template`, func(e *colly.HTMLElement) {
		document, err := goquery.NewDocumentFromReader(strings.NewReader(e.Text))
		if err != nil {
			panic(err)
		}

		document.Find("h3").Each(func(i int, selection *goquery.Selection) {

			val, exists := selection.Contents().Attr("href")
			if !exists {
				panic("Attr href doesn't exist")
			}
			category := strings.Replace(val, "#", "", 1)
			category = strings.Replace(category, "configs", "", 1)

			node := selection.NextAllFiltered(".data-table").Get(0)

			s2 := goquery.Selection{}
			selection2 := s2.AddNodes(node)

			columns := make(map[int]string)

			selection2.Find("tr").Each(func(i int, selection *goquery.Selection) {

				fmt.Println(selection.Text())
				if i == 0 {
					selection.Find("th").Each(func(j int, selection *goquery.Selection) {
						fmt.Println(selection.Text())
						columns[j] = selection.Text()
					})
					return
				}

				p := property{}
				p.Category = category

				selection.Find("td").Each(func(j int, selection *goquery.Selection) {
					column := columns[j]

					switch column {
					case "Name":
						p.Name = selection.Text()
					case "Description":
						p.Description = selection.Text()
					case "Type":
						p.Type = selection.Text()
					case "Default":
						p.Default = selection.Text()
					case "Valid Values":
						p.ValidValues = selection.Text()
					case "Importance":
						p.Importance = selection.Text()
					case "Dynamic Update Mode":
						p.DynamicUpdateMode = selection.Text()
					case "Server Default Property":
						p.ServerDefaultProperty = selection.Text()
					}

				})

				properties = append(properties, p)
			})

		})

	})
	c.OnError(func(response *colly.Response, e error) {
		if response.StatusCode == 404 {
			panic(errors.Errorf("Not found"))
		}
		panic(errors.Wrapf(e, "colly error (status:%d)", response.StatusCode))
	})
	err := c.Visit("https://kafka.apache.org/" + versionCode + "/documentation.html")
	if err != nil {
		panic(err)
	}

	return properties
}
