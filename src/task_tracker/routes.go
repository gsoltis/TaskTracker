package task_tracker

import (
	"net/http"
	"html/template"
	"appengine"
	"fmt"
)

type TaskTrackerUser struct {}

func init() {
	http.HandleFunc("/_ah/start", startup)
	http.Handle("/", authMiddleware(root))
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

func authMiddleware(next UserAwareHttpHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := authRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			appengine.NewContext(r).Debugf("User? %v", user)
			next(user, w, r)
		}
	})
}

var templates = template.Must(template.ParseGlob("templates/*"))



type MainPage struct {
	User *TaskTrackerUser
}

func root(user *TaskTrackerUser, w http.ResponseWriter, req *http.Request) {
	mp := MainPage{
		User: user,
	}
	templates.ExecuteTemplate(w, "body", &mp)
}
