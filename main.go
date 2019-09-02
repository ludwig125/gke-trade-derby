package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/sclevine/agouti"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	number             = flag.Int("number", 12345, "env number.")
	debug              = flag.Bool("debug", false, "Use Debug Mode?")
	debugUser          = flag.String("debuguser", "", "Debug User")
	debugSheetID       = flag.String("debugsheetid", "", "Debug SheetID")
	user               = "noset"
	pass               = "noset"
	sheetID            = ""
	credentialFileDir  = "credential"
	credentialFilePath = ""
)

type stockInfo struct {
	Code   string
	Name   string
	Status string
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s environment variable not set.", k)
	}
	log.Printf("%s environment variable set.", k)
	//log.Println("this is env", k, ":", v)
	return v
}

func fileMustExists(name string) string {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		log.Fatalf("file '%s' does not exists", name)
	}
	return name
}

func init() {
	flag.Parse()

	log.Printf("use debug mode?: %t", *debug)
	if *debug {
		// localで実行する場合はdebug modeを利用してリアクティブに実行する
		// debug modeの場合はdebuguserをオプションで指定し、passwordを入力する

		u := *debugUser
		if u == "" {
			log.Fatal("debugUser noset. if you use debug=true, set debugUser")
		}
		user = u

		s := *debugSheetID
		if s == "" {
			log.Fatal("debugSheetID noset. if you use debug=true, set debugSheetID")
		}
		sheetID = s

		fmt.Print("Password: ")
		p, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal("Failed to read password", err)
		}
		pass = string(p)
		log.Println("set debug infos")
	} else {
		// testまたはGKEから実行する場合は環境変数から取得する
		user = mustGetenv("APPUSER")
		pass = mustGetenv("APPPASS")
		sheetID = mustGetenv("TRADEDERBY_SHEETID")
	}

	// testまたはdebug用
	cdir := os.Getenv("CREDENTIALFILE_DIR")
	if cdir != "" {
		credentialFileDir = cdir
	}
	// Sheetをみにいくためのserviceaccountの置き場所
	// ローカルでもGKEでも同じになるようにvolumeのmountPathなどを調節している
	credentialFilePath = fileMustExists(credentialFileDir + "/gke-trade-derby-serviceaccount.json")

	log.Println("set env infos")
}

func main() {
	// use PORT environment variable, or default to 8080
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	server := http.NewServeMux()
	server.HandleFunc("/", indexHandler)
	server.HandleFunc("/tradederby", tradeDerby)
	server.HandleFunc("/tradederby2", tradeDerby2)
	server.HandleFunc("/tradederby3", tradeDerby3)
	server.HandleFunc("/tradederby4", tradeDerby4)
	server.HandleFunc("/tradederby5", tradeDerby5)

	// start the web server on port and accept requests
	log.Printf("Server listening on port: %s", port)
	err := http.ListenAndServe(":"+port, server)
	log.Fatal(err)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request: %s", r.URL.Path)
	host, _ := os.Hostname()
	fmt.Fprintf(w, "trade derby\n")
	fmt.Fprintf(w, "Hostname: %s\n", host)
	fmt.Fprintf(w, "cpu: %d\n", runtime.NumCPU())
	fmt.Fprintf(w, "GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
}

func tradeDerby(w http.ResponseWriter, r *http.Request) {
	html, err := fetchStockDocFromWebPage(user, pass)
	if err != nil {
		log.Fatalf("Failed to fetchStockDocFromWebPage, %v", err)
	}

	stockInfos, err := fetchStockInfo(html)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stockInfos)

	// stockInfos := []stockInfo{
	// 	stockInfo{"1417", "ミライトHD", "信用売"},
	// 	stockInfo{"6088", "シグマクシス", "現物買"},
	// 	stockInfo{"6367", "ダイキン", "現物買"},
	// 	stockInfo{"9759", "NSD", "信用売"},
	// }

	var sIfs [][]interface{}
	for _, s := range stockInfos {
		var sIf []interface{}
		sIf = append(sIf, s.Code)
		sIf = append(sIf, s.Name)
		sIf = append(sIf, s.Status)
		sIfs = append(sIfs, sIf)
	}
	log.Println(sIfs)

	// spreadsheetのclientを取得
	srv, err := getSheetClient()
	if err != nil {
		log.Fatalf("failed to get sheet client. err: %v", err)
	}
	log.Println("succeeded to get sheet client")

	log.Println("trying to write sheet")
	if err := clearAndWriteSheet(srv, sheetID, "trade-derby", sIfs); err != nil {
		log.Fatalf("failed to clearAndWriteSheet. %v", err)
	}
	log.Println("succeeded to write sheet")
}

func tradeDerby2(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Failed to start driver: %v", err)
	}
	defer driver.Stop()
	log.Println("succeeded to start WebDriver")

	// WebDriverの新規セッションを作成
	page, err := driver.NewPage()
	if err != nil {
		log.Printf("Failed to open page: %v", err)
	}
	log.Println("succeeded to start new WebDriver session")

	loginURL := "https://www.k-zone.co.jp/td/users/login"
	if err := page.Navigate(loginURL); err != nil {
		log.Printf("failed to navigate: %v", err)
	}

	time.Sleep(1 * time.Second)

	count(page, "tradederby2")
	log.Println("finished successfully")
}

func tradeDerby3(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Failed to start driver: %v", err)
	}
	defer driver.Stop()
	log.Println("succeeded to start WebDriver")

	// WebDriverの新規セッションを作成
	page, err := driver.NewPage()
	if err != nil {
		log.Printf("Failed to open page: %v", err)
	}
	log.Println("succeeded to start new WebDriver session")

	loginURL := "https://www.k-zone.co.jp/td/users/login"
	if err := page.Navigate(loginURL); err != nil {
		log.Printf("failed to navigate: %v", err)
	}

	time.Sleep(1 * time.Second)

	count(page, "tradederby3-1")

	// IDの要素を取得し、値を設定
	identity := page.FindByID("login_id")
	if err := identity.Fill(user); err != nil {
		log.Fatalf("failed to Fill login_id: %v", err)
	}
	log.Println("succeeded to fill login")

	count(page, "tradederby3-2")
	log.Println("finished successfully")
}

func tradeDerby4(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Failed to start driver: %v", err)
	}
	defer driver.Stop()
	log.Println("succeeded to start WebDriver")

	// WebDriverの新規セッションを作成
	page, err := driver.NewPage()
	if err != nil {
		log.Printf("Failed to open page: %v", err)
	}
	log.Println("succeeded to start new WebDriver session")

	loginURL := "https://www.k-zone.co.jp/td/users/login"
	if err := page.Navigate(loginURL); err != nil {
		log.Printf("failed to navigate: %v", err)
	}

	time.Sleep(1 * time.Second)

	count(page, "tradederby4-1")

	// IDの要素を取得し、値を設定
	identity := page.FindByID("login_id")
	if err := identity.Fill(user); err != nil {
		log.Fatalf("failed to Fill login_id: %v", err)
	}
	log.Println("succeeded to fill login")

	count(page, "tradederby4-2")

	// passwordの要素を取得し、値を設定
	password := page.FindByName("password")
	if err := password.Fill(pass); err != nil {
		log.Fatalf("failed to Fill login_id: %v", err)
	}
	log.Println("succeeded to fill pass")

	count(page, "tradederby4-3")

	// if err := page.FindByID("login_button").Submit(); err != nil {
	// 	log.Fatalf("failed to confirm password: %v", err)
	// }
	//count(page, "tradederby4-4")

	log.Println("finished successfully")
}

func tradeDerby5(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Failed to start driver: %v", err)
	}
	defer driver.Stop()
	log.Println("succeeded to start WebDriver")

	// WebDriverの新規セッションを作成
	page, err := driver.NewPage()
	if err != nil {
		log.Printf("Failed to open page: %v", err)
	}
	log.Println("succeeded to start new WebDriver session")

	loginURL := "https://www.k-zone.co.jp/td/dashboards/position_hold?lang=ja"
	if err := page.Navigate(loginURL); err != nil {
		log.Printf("failed to navigate: %v", err)
	}

	time.Sleep(1 * time.Second)

	count(page, "tradederby5-1")

	// IDの要素を取得し、値を設定
	identity := page.FindByID("login_id")
	if err := identity.Fill(user); err != nil {
		log.Fatalf("failed to Fill login_id: %v", err)
	}
	log.Println("succeeded to fill login")

	count(page, "tradederby5-2")

	// passwordの要素を取得し、値を設定
	password := page.FindByName("password")
	if err := password.Fill(pass); err != nil {
		log.Fatalf("failed to Fill login_id: %v", err)
	}
	log.Println("succeeded to fill pass")

	count(page, "tradederby5-3")

	if err := page.FindByID("login_button").Submit(); err != nil {
		//log.Fatalf("failed to confirm password: %v", err)
		log.Println("failed to confirm password")
	}
	count(page, "tradederby5-4")

	stockInfoURL := "https://www.k-zone.co.jp/td/dashboards/position_hold?lang=ja"
	if err := page.Navigate(stockInfoURL); err != nil {
		log.Printf("Failed to navigate bookstore page: %v", err)
	}
	count(page, "tradederby5-5")

	log.Println("finished successfully")
}

func count(page *agouti.Page, str string) {
	log.Println(str, "find id")
	file := fmt.Sprintf("/tmp/%s.jpg", str)
	page.Screenshot(file)
	makeHTML(page, str)

	s := page.FindByID("login_button")
	//log.Printf("selection --'%#v'--, --'%v'--\n\n", sele, sele)
	//log.Printf("%T\n", s)
	cnt, err := s.Count()
	if err != nil {
		log.Printf("failed to select elements from %s: %v", s, err)
		return
	}
	log.Println("len ele", cnt)
}

func makeHTML(page *agouti.Page, str string) {
	html, err := page.HTML()
	if err != nil {
		log.Printf("failed to get HTML: %v", err)
	}

	path := fmt.Sprintf("/tmp/%s.html", str)
	file, err := os.Create(path)
	if err != nil {
		log.Printf("failed to open file %s", path)
	}
	defer file.Close()

	file.Write(([]byte)(html))
}
