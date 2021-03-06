package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"

	_ "github.com/lib/pq"
)

// BaseURL is the Url of the website
const BaseURL = "gorl.herokuapp.com/"

var (
	repeat int
	db     *sql.DB
)

// Page is a Home page template
type Page struct {
	Info     string
	Original string
	Short    string
}

var templates = template.Must(template.ParseFiles("./view/home.html"))

func renderTemplate(w http.ResponseWriter, p *Page) {
	err := templates.ExecuteTemplate(w, "home.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/([-a-zA-Z0-9@:%_+.~#?&=]+)?/?$")
var validShort = regexp.MustCompile("^[-a-zA-Z0-9@:%_+.~#?&=]+$")
var validOriginal = regexp.MustCompile("^(https?://)(www[.])?[-a-zA-Z0-9@:%._+~#=]{2,256}[.][a-z]{2,4}([-a-zA-Z0-9@:%_+.~#?&/=]*)$")

func determineListenAddress() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}

func home(w http.ResponseWriter, r *http.Request) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		// Invalid path
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if m[1] == "" {
		// Home path or URL checking
		info := "Welcome, here you can shorten any URL."
		short := r.FormValue("short")
		original := r.FormValue("original")

		if short != "" {
			// A short URL is given
			// Check if short is in the database
			redirect, _ := getOriginal(short)
			if redirect != "" {
				// Short URL is already in the database
				info = BaseURL + short + " is already taken, please try something else."
			} else if original != "" {
				// Short URL is free and an URL to redirect is given
				// Add these to the database
				aerr := addOriginal(short, original)
				if aerr != nil {
					info = aerr.Error()
				} else {
					info = BaseURL + short + " will now be redirected to " + original
					short = ""
					original = ""
				}
			} else {
				// There is URL to redirect to
				info = "Where would you like to redirect " + BaseURL + short + " ?"
			}
		} else {
			// No short URL is given
			// @TODO generate a random valid short URL
			short = generateShort()
		}
		renderTemplate(w, &Page{Info: info, Original: original, Short: short})
		return
	}
	// Check if m[1] exists in the database
	redirect, _ := getOriginal(m[1])
	if redirect != "" {
		// The url is in the database
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}
	// The URL is free
	http.Redirect(w, r, "/?"+"short"+"="+m[1], http.StatusFound)
	return
}

func generateShort() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for n := 2; n < 11; n++ {
		// Generating a bigger and bigger url
		b := make([]byte, n)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		if short, _ := getOriginal(string(b)); short == "" {
			// Checking if the url is valid
			return string(b)
		}
	}
	return ""
}

func getOriginal(short string) (string, error) {
	if !validShort.MatchString(short) {
		return "", fmt.Errorf("Unvalid short url")
	}
	var (
		original string
	)
	err := db.QueryRow("SELECT short, original FROM urls WHERE short LIKE "+"'"+short+"'").Scan(&short, &original)
	if err != nil {
		return "", err
	}
	return original, nil
}

func addOriginal(short string, original string) error {
	if !validShort.MatchString(short) {
		return fmt.Errorf("Unvalid short url")
	}
	if !validOriginal.MatchString(original) {
		return fmt.Errorf("Unvalid original url")
	}
	if _, err := db.Exec("INSERT INTO urls (short, original) VALUES ( '" + short + "','" + original + "')"); err != nil {
		return err
	}
	return nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	addr, err := determineListenAddress()
	if err != nil {
		log.Fatal(err)
	}

	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS urls (short VARCHAR, original VARCHAR)"); err != nil {
		log.Fatalf("Error creating table: %q", err)
	}

	http.HandleFunc("/", home)

	log.Printf("Listening on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
