package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func determineListenAddress() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}

func hello(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	if url == "/" {
		fmt.Fprintln(w, "Welcome on the url shortene !")
	} else {
		fmt.Fprintln(w, "The url "+url+" is free, where would you want to redirect ?")
	}
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
}
