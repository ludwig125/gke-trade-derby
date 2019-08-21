package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"syscall"

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

	// Sheetをみにいくためのserviceaccountの置き場所
	// ローカルでもGKEでも同じになるようにvolumeのmountPathなどを調節している
	credentialFilePath = fileMustExists("credential/gke-trade-derby-serviceaccount.json")

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
		log.Println("set env infos")
	}

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
