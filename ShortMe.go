/*
** ShortMe.go
** Author: Marin Alcaraz
** Mail   <marin.alcaraz@gmail.com>
** Started on  Fri Apr 10 17:39:34 2015 Marin Alcaraz
** Last update Wed Apr 15 12:13:13 2015 Marin Alcaraz
 */

package main

import (
	"log"
	"math/rand"
	"net/http"
	"path"
	"regexp"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

/*
To do:
	RedirectAndCount

Nice to have:
	Besides random string use a user defined string
*/

//Config variables
var hostURL = "http://127.0.0.1:8080/"
var dbUsername = "marin"
var dbPass = "devel"
var dbName = "shorturl"
var dbEngine = "postgres"

//Globals
var db gorm.DB
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var maxShortSize = 10

//CustomURLs provides the model to interact with our DB
type customurls struct {
	ID        uint   `gorm:"primary_key"`
	OldURL    string `sql:"not null;unique;size:255"`
	ShortURL  string `sql:"not null;unique;size:255"`
	Visits    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

//Page provides an endpoint of valuable information
//needed by the fronted
type Page struct {
	Title    string
	NewURL   string
	Messages string
}

func generateShortURL(original string) string {
	b := make([]rune, maxShortSize)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return hostURL + string(b)
}

func validURL(url string) bool {
	r := regexp.MustCompile("(http(s*))://([\\w]*.)*")
	return r.MatchString(url)
}

func checkDuplicate(url string) string {
	shortItem := customurls{}
	db.Where("old_URL = ?", url).Find(&shortItem)
	return shortItem.ShortURL
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	lp := path.Join("templates", "index.html")
	index := Page{Title: "URLShortener - By Marin Alcaraz"}

	t, err := template.ParseFiles(lp)
	if err != nil {
		log.Fatal(err)
	}

	if r.Method == "POST" {
		shortItem := customurls{}
		r.ParseForm()
		shortItem.OldURL = r.FormValue("oldURL")
		if validURL(shortItem.OldURL) {
			s := generateShortURL(shortItem.OldURL)
			shortItem.ShortURL = s

			//Prevent duplicates
			exists := checkDuplicate(shortItem.OldURL)
			if exists == "" {
				db.NewRecord(shortItem)
				db.Create(&shortItem)
				index.NewURL = s
			} else {
				index.NewURL = exists
			}
		} else {
			index.Messages = "Invalid URL provided"
		}
	}
	t.Execute(w, index)
}

func initWebServer() {
	fs := http.FileServer(http.Dir("static"))

	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", indexHandler)

	log.Println("Listening...")
	http.ListenAndServe(":8080", nil)
}

func connectToDB() gorm.DB {
	dbConnectionString := "user=" + dbUsername +
		" password=" + dbPass +
		" dbname=" + dbName +
		" sslmode=disable"

	db, err := gorm.Open(dbEngine, dbConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	//Ping and configure the DB
	db.DB()
	db.DB().Ping()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	// Drop table
	db.DropTable(&customurls{})
	db.CreateTable((&customurls{}))

	return db
}

func main() {
	db = connectToDB()
	initWebServer()
}
