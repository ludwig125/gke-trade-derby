package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
)

var (
	number  = flag.Int("number", 12345, "env number.")
	envData = "default"
	user    = "noset"
	pass    = "noset"
)

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s environment variable not set.", k)
	}
	log.Printf("%s environment variable set.", k)
	return v
}

func init() {
	user = mustGetenv("APPUSER")
	pass = mustGetenv("APPPASS")
}

func main() {
	// use PORT environment variable, or default to 8080
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	if fromEnv := os.Getenv("ENVDATA"); fromEnv != "" {
		envData = fromEnv
	}
	log.Printf("this is ENVDATA '%s'", envData)

	log.Printf("this is USER '%s'", user)
	log.Printf("this is PASS '%s'", pass)

	server := http.NewServeMux()
	server.HandleFunc("/", indexHandler)

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

	flag.Parse()
	fmt.Fprintf(w, "ENV NUMBER: %d\n", *number)
	fmt.Fprintf(w, "ENV DATA: %s\n", envData)
}
