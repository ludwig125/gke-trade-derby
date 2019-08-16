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
	number    = flag.Int("number", 12345, "env number.")
	debug     = flag.Bool("debug", false, "Use Debug Mode?")
	debugUser = flag.String("debuguser", "", "Debug User")
	user      = "noset"
	pass      = "noset"
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
	return v
}

func init() {
	flag.Parse()

	log.Printf("use debug mode?: %t", *debug)
	if *debug {
		// debug modeの場合はdebuguserをオプションで指定し、passwordを入力する
		u := *debugUser
		if u == "" {
			log.Fatal("debuguser noset. if you use debug=true, set debuguser")
		}
		user = u
		log.Println("set debuguser:", u)

		fmt.Print("Password: ")
		p, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal("Failed to read password", err)
		}
		pass = string(p)
		log.Println("set debugpass")
	} else {
		// GKEから実行する場合は環境変数から取得する
		user = mustGetenv("APPUSER")
		pass = mustGetenv("APPPASS")
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
}
