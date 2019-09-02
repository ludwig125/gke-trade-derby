package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sclevine/agouti"
)

// use agouti
// https://godoc.org/github.com/sclevine/agouti
func fetchStockDocFromWebPage(user string, pass string) (string, error) {
	// ブラウザはChromeを指定して起動
	//driver := agouti.ChromeDriver(agouti.Browser("chrome"))
	driver := agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless", // ブラウザを立ち上げないheadlessモードの指定
			//"--window-size=1280,800", // ウィンドウサイズの指定
			"--disable-gpu", // 暫定的に必要なフラグ
			"--no-sandbox",
		}),
		agouti.Debug,
	)
	if err := driver.Start(); err != nil {
		return "", fmt.Errorf("Failed to start driver: %v", err)
	}
	defer driver.Stop()
	log.Println("succeeded to start WebDriver")

	// https://godoc.org/github.com/sclevine/agouti
	// NewPage returns a *Page that corresponds to a new WebDriver session. Provided Options configure the page. For instance, to disable JavaScript:
	// WebDriverの新規セッションを作成
	page, err := driver.NewPage()
	if err != nil {
		return "", fmt.Errorf("Failed to open page: %v", err)
	}
	log.Println("succeeded to start new WebDriver session")

	loginURL := "https://www.k-zone.co.jp/td/users/login"
	if err := login(page, user, pass, loginURL); err != nil {
		return "", fmt.Errorf("failed to login, %v", err)
	}
	log.Println("succeeded to login")

	// 所有している株一覧のページに遷移してHTMLを取得
	stockInfoURL := "https://www.k-zone.co.jp/td/dashboards/position_hold?lang=ja"
	html, err := fetchStockDoc(page, stockInfoURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetchStockDoc, %v", err)
	}
	log.Println("succeeded to fetchStockDoc")
	return html, nil
}

func login(page *agouti.Page, user string, pass string, loginURL string) error {
	// ログインページに遷移
	if err := page.Navigate(loginURL); err != nil {
		return fmt.Errorf("failed to navigate: %v", err)
	}

	html, err := page.HTML()
	if err != nil {
		return fmt.Errorf("failed to get HTML: %v", err)
	}
	log.Println("---------------------------------------------------")
	log.Println("HTML:", html)
	log.Println("---------------------------------------------------")

	// HTML: view-source:https://www.k-zone.co.jp/td/users/login

	// IDの要素を取得し、値を設定
	identity := page.FindByID("login_id")
	if err := identity.Fill(user); err != nil {
		return fmt.Errorf("failed to Fill login_id: %v", err)
	}

	// passwordの要素を取得し、値を設定
	password := page.FindByName("password")
	if err := password.Fill(pass); err != nil {
		return fmt.Errorf("failed to Fill login_id: %v", err)
	}

	time.Sleep(1 * time.Second)

	// cnt, err := page.All("ログイン").Count()
	// if err != nil {
	// 	log.Printf("failed to count ログイン: %v", err)
	// }
	// log.Println("count ログイン", cnt)
	resFindByIDLoginButton := page.FindByID("login_button")
	log.Printf("resFindByIDLoginButton: %#v", resFindByIDLoginButton)

	log.Println("find id")
	sele := page.FindByID("login_button")
	log.Printf("selection --'%#v'--, --'%v'--\n\n", sele, sele)

	log.Printf("%T\n", sele)
	cnt, err := sele.Count()
	if err != nil {
		log.Fatalf("failed to select elements from %s: '%v'", sele, err)
	}
	log.Println("len ele", cnt)

	if err := page.FindByID("login_button").Submit(); err != nil {
		return fmt.Errorf("failed to confirm password: %v", err)
	}
	return nil
}

func fetchStockDoc(page *agouti.Page, stockInfoURL string) (string, error) {
	if err := page.Navigate(stockInfoURL); err != nil {
		return "", fmt.Errorf("Failed to navigate bookstore page: %v", err)
	}
	//time.Sleep(1 * time.Second)

	html, err := page.HTML()
	if err != nil {
		return "", fmt.Errorf("Failed to get html: %v", err)
	}
	return html, nil
}
