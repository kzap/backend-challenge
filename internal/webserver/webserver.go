package webserver

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kzap/ada-backend-challenge/internal/config"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// Route describes the URLs available for this program
type Route struct {
	URL         string
	Description string
	Methods     string
	Handler     func(w http.ResponseWriter, r *http.Request)
}

type Conversation struct {
	ID             int    `json:"-"`
	ConversationID string `json:"-"`
	Sender         string `json:"sender"`
	Message        string `json:"message"`
	CreatedDate    string `json:"created"`
}

var templates = template.Must(template.ParseGlob(getCallerDir() + "/../../web/template/*.html"))

var routes []Route
var db *sql.DB

const webPort = 8000

// Start starts the HTTP Server
func Start(mysqlConfig config.DbConfig) {
	var err error

	routes = append(routes, Route{URL: "/", Description: "Homepage", Methods: "GET", Handler: GetIndex})
	routes = append(routes, Route{URL: "/conversations", Description: "Show conversations", Methods: "GET", Handler: GetConversationList})
	routes = append(routes, Route{URL: "/conversations/{conversation_id}.json", Description: "View a Conversation", Methods: "GET", Handler: GetConversationJSON})
	routes = append(routes, Route{URL: "/conversations/{conversation_id}", Description: "View a Conversation", Methods: "GET", Handler: GetConversation})
	routes = append(routes, Route{URL: "/messages", Description: "Message Input Form", Methods: "GET", Handler: GetMessages})
	routes = append(routes, Route{URL: "/messages", Description: "Message POST Handler", Methods: "POST", Handler: PostMessages})

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

func getConversationsByID(conversationID string) ([]Conversation, error) {
	var err error
	conversations := []Conversation{}

	rows, err := db.Query("SELECT `id`, `conversation_id`, `sender`, `message`, `created_date` FROM `conversations` WHERE `conversation_id` = ?", conversationID)
	if err != nil {
		log.Println(err)
		return conversations, errors.New("ERROR_DB_QUERY")
	}
	defer rows.Close()

	for rows.Next() {
		var c Conversation
		err = rows.Scan(&c.ID, &c.ConversationID, &c.Sender, &c.Message, &c.CreatedDate)
		if err != nil {
			log.Println(err)
			return conversations, errors.New("ERROR_ROWS_SCAN")
		}
		conversations = append(conversations, c)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return conversations, errors.New("ERROR_ROWS_NEXT")
	}

	if len(conversations) == 0 {
		log.Println("Invalid Conversation ID")
		return conversations, errors.New("CONVERSATION_NOT_FOUND")
	}

	return conversations, nil
}

// GetConversation shows the conversation
func GetConversation(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)

	conversationID := vars["conversation_id"]
	conversations, err := getConversationsByID(conversationID)
	if err != nil {
		renderError(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
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

// GetConversationJSON shows the conversation results as JSON
func GetConversationJSON(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)

	conversationID := vars["conversation_id"]
	conversations, err := getConversationsByID(conversationID)
	if err != nil {
		renderJSONError(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	data := struct {
		ConversationID string         `json:"id"`
		Conversations  []Conversation `json:"messages"`
	}{
		ConversationID: conversationID,
		Conversations:  conversations,
	}
	json.NewEncoder(w).Encode(data)
}

// GetMessages shows the Message form
func GetMessages(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	err := templates.ExecuteTemplate(w, "messages.html", nil)
	if err != nil {
		renderError(w, "EXECUTE_TEMPLATE_ERROR", http.StatusInternalServerError)
		log.Println(err)
	}
}

// PostMessages handles the receiving of messages
func PostMessages(w http.ResponseWriter, r *http.Request) {
	var err error

	formConversationID := r.FormValue("conversation_id")
	if formConversationID == "" {
		formConversationID = randToken(8)
	}
	formSender := r.FormValue("sender")
	if strings.TrimSpace(formSender) == "" {
		renderError(w, "FORM_EMPTY_SENDER", http.StatusUnprocessableEntity)
		return
	}
	formMessage := r.FormValue("message")
	if strings.TrimSpace(formMessage) == "" {
		renderError(w, "FORM_EMPTY_MESSAGE", http.StatusUnprocessableEntity)
		return
	}
	formRedirect := r.FormValue("redirect")
	redirect := false
	if formRedirect != "" {
		redirect, err = strconv.ParseBool(formRedirect)
		if err != nil {
			renderError(w, "ERROR_PARSEBOOL", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	// perform a db.Query insert
	insertQuery := fmt.Sprintf(
		"INSERT INTO `conversations` VALUES (DEFAULT, '%v', '%v', '%v', DEFAULT)",
		formConversationID,
		formSender,
		formMessage)
	_, err = db.Query(insertQuery)

	// if there is an error inserting, handle it
	if err != nil {
		renderError(w, "CANT_INSERT_DB", http.StatusInternalServerError)
		log.Println(err)
		log.Println(insertQuery)
		return
	}

	if redirect {
		http.Redirect(w, r, "/conversations/"+formConversationID, 302)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("SUCCESS"))
}

func renderError(w http.ResponseWriter, errorMsg string, responseCode int) {
	w.WriteHeader(responseCode)
	w.Write([]byte(errorMsg))
}

func renderJSONError(w http.ResponseWriter, errorMsg string, responseCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)

	data := struct {
		Status int    `json:"status"`
		Error  string `json:"error"`
	}{
		Status: responseCode,
		Error:  errorMsg,
	}
	json.NewEncoder(w).Encode(data)
}

func randToken(len int) string {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("read.Read error", err)
	}

	return fmt.Sprintf("%x", b)
}

func getCallerDir() string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}

	return path.Dir(filename)
}
