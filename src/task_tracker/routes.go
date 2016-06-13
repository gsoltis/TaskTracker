package task_tracker

import (
	"github.com/gorilla/mux"
	"net/http"
	"html/template"
	"appengine"
	"fmt"
)



func init() {
	r := mux.NewRouter()
	r.HandleFunc("/_ah/start", startup)
	r.HandleFunc("/login", authRequest)
	r.HandleFunc("/", root)
	http.Handle("/", r)
}

func startup(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("HAndling startup")
	err := InitKeyCache(appengine.NewContext(req))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w, "Ok")
	}
}

type UserAwareHttpHandler func(*TaskTrackerUser, http.ResponseWriter, *http.Request)

var templates = template.Must(template.ParseGlob("templates/*"))



type MainPage struct {
	User *TaskTrackerUser
}

func root(w http.ResponseWriter, req *http.Request) {
	user, err := UserForRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	mp := MainPage{
		User: user,
	}
	templates.ExecuteTemplate(w, "body", &mp)
}
