package webserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// GetIndex shows the homepage
func TestGetIndex(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetIndex)

	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestGetConversationList(t *testing.T) {
	/*
		req, err := http.NewRequest("GET", "/conversations", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetConversationList)

		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	*/
}

func TestgetConversationsByID(t *testing.T) {

}

// GetConversation shows the conversation
func TestGetConversation(t *testing.T) {

}

// GetConversationJSON shows the conversation results as JSON
func TestGetConversationJSON(t *testing.T) {

}

// PostMessages handles the receiving of messages
func TestPostMessages(t *testing.T) {

}

func TestrenderError(t *testing.T) {

}

func TestrenderJSONError(t *testing.T) {

}

func TestrandToken(t *testing.T) {

}
