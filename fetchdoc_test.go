package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sclevine/agouti"
)

func dummyLoginHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, _ := ioutil.ReadFile("login.html")
		fmt.Fprintf(w, string(content))
	})
}

func dummyStockHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, _ := ioutil.ReadFile("stock.html")
		fmt.Fprintf(w, string(content))
	})
}

func TestFetchStockDocFromWebPage(t *testing.T) {

	driver := agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless", // ブラウザを立ち上げないheadlessモードの指定
		}),
		agouti.Debug,
	)
	if err := driver.Start(); err != nil {
		t.Fatalf("Failed to start driver: %v", err)
	}
	defer driver.Stop()

	page, err := driver.NewPage()
	if err != nil {
		t.Fatalf("Failed to open page: %v", err)
	}

	t.Run("testLogin", func(t *testing.T) {
		testLogin(t, page)
	})
	t.Run("testFetchStockDoc", func(t *testing.T) {
		testFetchStockDoc(t, page)
	})

}

func testLogin(t *testing.T, page *agouti.Page) {
	ts := httptest.NewServer(dummyLoginHandler())
	defer ts.Close()

	if err := login(page, "user", "pass", ts.URL); err != nil {
		t.Error(err)
	}
}

func testFetchStockDoc(t *testing.T, page *agouti.Page) {
	ts := httptest.NewServer(dummyStockHandler())
	defer ts.Close()

	_, err := fetchStockDoc(page, ts.URL)
	if err != nil {
		t.Error(err)
	}
}
