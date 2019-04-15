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
	Methods     string `json:"methods,omitempty"`
	Handler     func(w http.ResponseWriter, r *http.Request)
}

type Conversation struct {
	ID             int
	ConversationID string
	Sender         string
	Message        string
	CreatedDate    string
}

// parse templates dir
var templates = template.Must(template.ParseGlob("./web/template/*.html"))

var routes []Route
var db *sql.DB

const webPort = 8000

// Start starts the HTTP Server
func Start(mysqlConfig config.DbConfig) {
	var err error

	routes = append(routes, Route{URL: "/", Description: "Homepage", Methods: "GET", Handler: GetIndex})
	routes = append(routes, Route{URL: "/conversations", Description: "Show conversations", Methods: "GET", Handler: GetConversationList})
	routes = append(routes, Route{URL: "/conversations/{conversation_id}", Description: "View a Conversation", Methods: "GET", Handler: GetConversation})
	routes = append(routes, Route{URL: "/messages", Description: "Message handler", Methods: "POST", Handler: PostMessages})

	router := mux.NewRouter()
	router.StrictSlash(true)

	for _, route := range routes {
		router.HandleFunc(route.URL, route.Handler).Methods(route.Methods)
	}
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
	err := templates.ExecuteTemplate(w, "index.html", routes)
	if err != nil {
		renderError(w, "EXECUTE_TEMPLATE_ERROR", http.StatusInternalServerError)
		log.Println(err)
	}
}

// GetConversationList shows the conversation
func GetConversationList(w http.ResponseWriter, r *http.Request) {
	var err error

	rows, err := db.Query("SELECT DISTINCT `conversation_id` FROM `conversations`")
	if err != nil {
		renderError(w, "ERROR_DB_QUERY", http.StatusInternalServerError)
		log.Println(err)

		return
	}
	defer rows.Close()

	conversations := []string{}
	for rows.Next() {
		var conversationID string
		err = rows.Scan(&conversationID)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				renderError(w, "ERROR_NOT_FOUND", http.StatusNotFound)
			default:
				renderError(w, "ERROR_ROWS_SCAN", http.StatusInternalServerError)
			}
			log.Println(err)
			return
		}
		conversations = append(conversations, conversationID)
	}
	err = rows.Err()
	if err != nil {
		renderError(w, "ERROR_ROWS_NEXT", http.StatusInternalServerError)
		log.Println(err)

		return
	}

	w.WriteHeader(http.StatusOK)

	err = templates.ExecuteTemplate(w, "conversation_list.html", conversations)
	if err != nil {
		renderError(w, "EXECUTE_TEMPLATE_ERROR", http.StatusInternalServerError)
		log.Println(err)
	}
}

// GetConversation shows the conversation
func GetConversation(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)

	conversationID := vars["conversation_id"]
	rows, err := db.Query("SELECT `id`, `conversation_id`, `sender`, `message`, `created_date` FROM `conversations` WHERE `conversation_id` = ?", conversationID)
	if err != nil {
		renderError(w, "ERROR_DB_QUERY", http.StatusInternalServerError)
		log.Println(err)

		return
	}
	defer rows.Close()

	conversations := []Conversation{}
	for rows.Next() {
		var c Conversation
		err = rows.Scan(&c.ID, &c.ConversationID, &c.Sender, &c.Message, &c.CreatedDate)
		if err != nil {
			renderError(w, "ERROR_ROWS_SCAN", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		conversations = append(conversations, c)
	}
	err = rows.Err()
	if err != nil {
		renderError(w, "ERROR_ROWS_NEXT", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if len(conversations) == 0 {
		renderError(w, "CONVERSATION_NOT_FOUND", http.StatusNotFound)
		log.Println("Invalid Conversation ID")
		return
	}

	w.WriteHeader(http.StatusOK)

	data := struct {
		ConversationID string
		Conversations  []Conversation
	}{
		ConversationID: conversationID,
		Conversations:  conversations,
	}
	err = templates.ExecuteTemplate(w, "conversation.html", data)
	if err != nil {
		renderError(w, "EXECUTE_TEMPLATE_ERROR", http.StatusInternalServerError)
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
