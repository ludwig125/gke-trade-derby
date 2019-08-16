package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// use goquery
// https://godoc.org/github.com/PuerkitoBio/goquery
func fetchStockInfo(html string) ([]stockInfo, error) {
	htmlReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		return []stockInfo{}, fmt.Errorf("Failed to load html doc. err: %v", err)
	}

	//func scrapeStockInfo() ([]stockInfo, error){
	var stockInfos []stockInfo
	// 一覧の株それぞれを見ていく
	doc.Find(".stockData").Each(func(_ int, s *goquery.Selection) {
		// upRowというタグに情報がある
		upRow := s.Find(".upRow")
		var codeName, status string
		upRow.Find("td").Each(func(i int, td *goquery.Selection) {
			if i == 0 { // 銘柄コードと企業名
				codeName = td.Text()
			} else if i == 1 { // 自分が買いで持っているか売りで持っているか
				status = td.Text()
			}
		})
		stockInfos = append(stockInfos, stockInfo{Code: codeName[:4], Name: codeName[4:], Status: status})
	})

	if stockInfos == nil {
		return []stockInfo{}, fmt.Errorf("failed to scrape stockInfo")
	}

	return stockInfos, nil
}
