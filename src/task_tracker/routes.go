package task_tracker

import (
	"github.com/gorilla/mux"
	"net/http"
	"html/template"
	"appengine"
	"fmt"
	"io/ioutil"
	"appengine/datastore"
	"strconv"
	"encoding/json"
)



func init() {
	r := mux.NewRouter()
	r.HandleFunc("/_ah/start", startup)
	r.HandleFunc("/login", authRequest)
	r.HandleFunc("/api/progress", progressHandler)
	r.HandleFunc("/api/tasks", addTask).Methods("POST")
	r.HandleFunc("/api/tasks", getTasks).Methods("GET")
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

var templates = template.Must(template.ParseGlob("templates/*"))

type Task struct {
	Name string
}

type Period struct {

}

type Goal struct {
	Task *Task
	Period *Period
}

type MainPage struct {
	User *TaskTrackerUser
}

func InternalServerError(w http.ResponseWriter, req *http.Request, err error) {
	var err_string = ""
	if err != nil {
		err_string = err.Error()
		appengine.NewContext(req).Errorf("Error: %v", err)
	}
	http.Error(w, err_string, http.StatusInternalServerError)
}

func RequireAuth(w http.ResponseWriter, req *http.Request) *TaskTrackerUser {
	user, err := UserForRequest(req)
	if err != nil {
		InternalServerError(w, req, err)
		return nil
	} else if user == nil {
		http.Error(w, "", http.StatusForbidden)
		return nil
	} else {
		return user
	}
}

func getTasks(w http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	user := RequireAuth(w, req)
	if user == nil {
		return
	}
	query := datastore.NewQuery("Task").Ancestor(user.Key(ctx)).Limit(10)
	tasks := make([]Task, 0, 10)
	keys, err := query.GetAll(ctx, &tasks)
	if err != nil {
		InternalServerError(w, req, err)
		return
	}
	key_map := make(map[string]*Task)
	for i, k := range keys {
		key_string := strconv.FormatInt(k.IntID(), 10)
		key_map[key_string] = &tasks[i]
	}
	json_bytes, err := json.Marshal(key_map)
	if err != nil {
		InternalServerError(w, req, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(json_bytes)
}

func addTask(w http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	user := RequireAuth(w, req)
	if user == nil {
		return
	}
	defer req.Body.Close()
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task_name := string(bytes)
	ctx.Debugf("Task name? %v", task_name)
	task := Task{task_name}


	user_key := user.Key(ctx)

	key := datastore.NewIncompleteKey(ctx, "Task", user_key)
	task_key, err := datastore.Put(ctx, key, &task)
	if err != nil {
		InternalServerError(w, req, err)
		return
	} else {
		ctx.Debugf("Task key? %v, %v", task_key.StringID(), task_key.IntID())
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte("\"" + strconv.FormatInt(task_key.IntID(), 10) + "\""))
	}

}

func progressHandler(w http.ResponseWriter, req *http.Request) {
	user := RequireAuth(w, req)
	if user == nil {
		return
	}

}

func root(w http.ResponseWriter, req *http.Request) {
	user, err := UserForRequest(req)
	if err != nil {
		InternalServerError(w, req, err)
		return
	}
	mp := MainPage{
		User: user,
	}
	templates.ExecuteTemplate(w, "body", &mp)
}
