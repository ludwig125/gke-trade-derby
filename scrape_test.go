package main

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestScrape(t *testing.T) {
	t.Run("testfetchStockInfoDoc", func(t *testing.T) {
		testFetchStockInfo(t)
	})
}

func testFetchStockInfo(t *testing.T) {
	b, _ := ioutil.ReadFile("stock.html")
	stockInfos, err := fetchStockInfo(string(b))
	if err != nil {
		t.Fatal(err)
	}

	//t.Log("stockInfos: ", stockInfos)
	wantStockInfos := []stockInfo{
		stockInfo{"1417", "ミライトHD", "信用売"},
		stockInfo{"6088", "シグマクシス", "現物買"},
		stockInfo{"6367", "ダイキン", "現物買"},
		stockInfo{"9759", "NSD", "信用売"},
	}

	if !reflect.DeepEqual(stockInfos, wantStockInfos) {
		t.Fatalf("scraped stockInfos: %#v, wantStockInfos: %#v", stockInfos, wantStockInfos)
	}
}
