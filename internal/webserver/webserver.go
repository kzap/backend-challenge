package webserver

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kzap/ada-backend-challenge/internal/config"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// Route describes the URLs available for this program
type Route struct {
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

// parse templates dir
var templates = template.Must(template.ParseGlob("./web/template/*.html"))

var routes []Route
var db *sql.DB

const webPort = 8000

// Start starts the HTTP Server
func Start(mysqlConfig config.DbConfig) {
	var err error

	routes = append(routes, Route{URL: "/", Description: "Index"})
	routes = append(routes, Route{URL: "/conversations/{conversation_id}", Description: "View a Conversation"})
	routes = append(routes, Route{URL: "/messages", Description: "Message handler"})

	router := mux.NewRouter()
	router.StrictSlash(true)

	router.HandleFunc("/", GetIndex).Methods("GET")
	router.HandleFunc("/conversations/{conversation_id}", GetConversation).Methods("GET")
	router.HandleFunc("/messages", PostMessages).Methods("POST")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static"))).Methods("GET")

	dbConnectionString := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", mysqlConfig.Username, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.DBName)
	log.Println(mysqlConfig, dbConnectionString)

	db, err = sql.Open("mysql", dbConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Web Server Starting on port [%v]...\n", webPort)
	// set timeouts for production settings
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", webPort), router))
}

// GetIndex shows the homepage
func GetIndex(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		renderError(w, "EXECUTE_TEMPLATE", http.StatusInternalServerError)
		log.Println(err)
	}
}

// GetConversation shows the conversation
func GetConversation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conversation := vars["conversation_id"]

	w.WriteHeader(http.StatusOK)
	err := templates.ExecuteTemplate(w, "conversation.html", conversation)
	if err != nil {
		renderError(w, "EXECUTE_TEMPLATE", http.StatusInternalServerError)
		log.Println(err)
	}
}

// PostMessages handles the receiving of messages
func PostMessages(w http.ResponseWriter, r *http.Request) {
	var err error

	// perform a db.Query insert
	insertQuery := fmt.Sprintf(
		"INSERT INTO `conversations` VALUES (DEFAULT, '%v', '%v', '%v', DEFAULT)",
		"conversation_id",
		"sender",
		"message")
	_, err = db.Query(insertQuery)
	// if there is an error inserting, handle it
	if err != nil {
		renderError(w, "CANT_INSERT_DB", http.StatusInternalServerError)
		log.Println(err)
		log.Println(insertQuery)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("SUCCESS"))
}

func renderError(w http.ResponseWriter, errorMsg string, responseCode int) {
	w.WriteHeader(responseCode)
	w.Write([]byte(errorMsg))
}
