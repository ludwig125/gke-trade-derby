package main

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestScrape(t *testing.T) {
	// fmt.Printf("cpu: %d\n", runtime.NumCPU())
	// fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
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

	t.Log("stockInfos: ", stockInfos)
	wantCode := []string{"1417", "6088", "6367", "6981", "7821", "9759", "9902"}
	var gotCode []string
	for _, v := range stockInfos {
		gotCode = append(gotCode, v.Code)
	}
	if !reflect.DeepEqual(gotCode, wantCode) {
		t.Fatalf("scraped code: %#v, want code: %#v", gotCode, wantCode)
	}
}
