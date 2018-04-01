package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"

	_ "github.com/lib/pq"
)

// DatabaseURL is the Url of the Postgres database
const DatabaseURL = "postgres://oezyzwrclcppmy:0508471cc1b64735ea793a6141c1872756b4c075c8ac521ee4681b855c5ea227@ec2-79-125-110-209.eu-west-1.compute.amazonaws.com:5432/dcqscah58liv58"
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

var validPath = regexp.MustCompile("^/([a-zA-Z0-9]+)?/?$")

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
			// @TODO Check if short is in the database
			redirect, _ := getOriginal(m[1])
			if redirect != "" {
				// Short URL is already in the database
				info = BaseURL + short + " is already taken, please try something else."
			} else if original != "" {
				// Short URL is free and an URL to redirect is given
				// @TODO Add these to the database
				info = BaseURL + short + " will now be redirected to " + original
			} else {
				// There is URL to redirect to
				info = "Where would you like to redirect " + BaseURL + short + " ?"
			}
		} else {
			// No short URL is given
			// @TODO generate a random short URL
			short = "A9oa"
		}
		renderTemplate(w, &Page{Info: info, Original: original, Short: short})
		return
	}
	// @TODO Check if m[1] exists in the database
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

func getOriginal(short string) (string, error) {
	var (
		original string
	)
	db, err := sql.Open("postgres", os.Getenv(DatabaseURL))
	if err != nil {
		return "", err
	}
	rows, err := db.Query("select short, original from urls where short = "+short, 1)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&short, &original)
		if err != nil {
			return "", err
		}
		return original, nil
	}
	err = rows.Err()
	if err != nil {
		return "", err
	}
	return "", nil
}

func main() {
	addr, err := determineListenAddress()
	if err != nil {
		log.Fatal(err)
	}

	db, err = sql.Open("postgres", os.Getenv(DatabaseURL))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS urls (short VARCHAR(),original VARCHAR())"); err != nil {
		log.Fatalf("Error creating table: %q", err)
	}

	http.HandleFunc("/", home)

	log.Printf("Listening on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
