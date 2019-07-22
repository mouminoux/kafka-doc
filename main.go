package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/pkg/errors"
	"os"
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
		"0101": "0.10.1.X",
		"0102": "0.10.2.X",
		"0110": "0.11.0.X",
		"10":   "1.0.X",
		"11":   "1.1.X",
		"20":   "2.0.X",
		"21":   "2.1.X",
		"22":   "2.2.X",
		"23":   "2.3.X",
	}

	for k, v := range kafkaVersion {
		getDoc(k, v)
	}
}

func getDoc(versionCode string, versionLabel string) {
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
