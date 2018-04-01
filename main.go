package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	_ "github.com/lib/pq"
)

// DatabaseURL is the Url of the Postgres database
const DatabaseURL = "postgres://oezyzwrclcppmy:0508471cc1b64735ea793a6141c1872756b4c075c8ac521ee4681b855c5ea227@ec2-79-125-110-209.eu-west-1.compute.amazonaws.com:5432/dcqscah58liv58"

var (
	repeat int
	db     *sql.DB
)

var validPath = regexp.MustCompile("^/([a-zA-Z0-9]+)/?$")

func determineListenAddress() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}

func hello(w http.ResponseWriter, r *http.Request) {
	m := validPath.FindStringSubmatch(r.URL.Path)

	if r.URL.Path == "/" || r.URL.Path == "" {
		fmt.Fprintln(w, "Welcome on the url shortener !")
		return
	} else if m == nil {
		//fmt.Fprintln(w, "You are being redirected ...")
		http.Redirect(w, r, "http://redirected.com", http.StatusFound)
		return
	}
	if false {
		fmt.Fprintln(w, "You are being redirected ...")
		http.Redirect(w, r, "http://redirected.com", http.StatusFound)
		return
	}
	fmt.Fprintln(w, "The url "+m[1]+" is free, where would you want to redirect ?")
	return
}
func main() {
	addr, err := determineListenAddress()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", hello)

	log.Printf("Listening on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}

	db, err = sql.Open("postgres", os.Getenv(DatabaseURL))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}
}
