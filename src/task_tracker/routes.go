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
	"errors"
	"github.com/fatih/structs"
)



func init() {
	r := mux.NewRouter()
	r.HandleFunc("/_ah/start", startup)
	r.HandleFunc("/login", authRequest)
	r.HandleFunc("/api/progress", progressHandler)
	r.HandleFunc("/api/tasks", addTask).Methods("POST")
	r.HandleFunc("/api/tasks", getTasks).Methods("GET")
	r.HandleFunc("/api/goals", getGoals).Methods("GET")
	r.HandleFunc("/api/goals", addGoal).Methods("POST")
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

type Goal struct {
	Task *datastore.Key
	Period Period
	Frequency int
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

func taskMapForKeys(ctx appengine.Context, keys []*datastore.Key) (map[string]interface{}, error) {
	tasks := make([]Task, len(keys), len(keys))
	err := datastore.GetMulti(ctx, keys, tasks)
	if err != nil {
		return nil, err
	}
	task_map := make(map[string]interface{})
	for i, k := range keys {
		task_map[strconv.FormatInt(k.IntID(), 10)] = structs.Map(tasks[i])
	}
	return task_map, nil
}

func taskMapForGoals(ctx appengine.Context, goals []Goal) (map[string]interface{}, error) {
	task_keys_map := make(map[*datastore.Key]*Task)
	var task_keys = make([]*datastore.Key, 0)
	for _, goal := range goals {
		if _, ok := task_keys_map[goal.Task]; !ok {
			task_keys_map[goal.Task] = nil
			task_keys = append(task_keys, goal.Task)
		}
	}
	return taskMapForKeys(ctx, task_keys)
}

func getGoals(w http.ResponseWriter, req *http.Request) {
	user := RequireAuth(w, req)
	if user == nil {
		return
	}
	ctx := appengine.NewContext(req)
	query := datastore.NewQuery("Goal").Ancestor(user.Key(ctx)).Limit(10)
	goals := make([]Goal, 0, 10)
	keys, err := query.GetAll(ctx, &goals)
	if err != nil {
		InternalServerError(w, req, err)
		return
	}
	tasks_map, err := taskMapForGoals(ctx, goals)
	if err != nil {
		InternalServerError(w, req, err)
		return
	}

	goals_map := make(map[string]interface{})
	for i, k := range keys {
		goal := goals[i]
		task_id := strconv.FormatInt(goal.Task.IntID(), 10)
		task_map := tasks_map[task_id]
		goal_map := structs.Map(goal)
		goal_map["Task"] = task_map
		goal_map["TaskId"] = task_id
		goals_map[strconv.FormatInt(k.IntID(), 10)] = goal_map
	}

	json_bytes, err := json.Marshal(goals_map)
	if err != nil {
		InternalServerError(w, req, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(json_bytes)

}

type ShallowGoal struct {
	TaskId string `json:"task_id"`
	Numerator int `json:"numerator"`
	Denominator string `json:"denominator"`
}

func (sg *ShallowGoal) TaskIntKey() (int64, error) {
	return strconv.ParseInt(sg.TaskId, 10, 64)
}

type Period int

const (
	Day Period = iota
	WorkDay
	Morning
	Evening
	WorkWeek
	Weekend
	Week
	Month
	Year
)

func PeriodFromString(s string) (Period, error) {
	switch s {
	case "week":
		return Week, nil
	default:
		return -1, errors.New("Unknown period")
	}
}

func addGoal(w http.ResponseWriter, req *http.Request) {
	user := RequireAuth(w, req)
	if user == nil {
		return
	}
	ctx := appengine.NewContext(req)
	defer req.Body.Close()
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var shallow_goal = &ShallowGoal{}
	err = json.Unmarshal(bytes, shallow_goal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	period, err := PeriodFromString(shallow_goal.Denominator)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx.Debugf("New goal: %v, period: %i", shallow_goal, period)
	task_id, err := shallow_goal.TaskIntKey()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user_key := user.Key(ctx)
	task_key := datastore.NewKey(ctx, "Task", "", task_id, user_key)
	goal := Goal{
		Task: task_key,
		Period: period,
		Frequency: shallow_goal.Numerator,
	}
	key := datastore.NewIncompleteKey(ctx, "Goal", user_key)
	goal_key, err := datastore.Put(ctx, key, &goal)
	if err != nil {
		InternalServerError(w, req, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte("\"" + strconv.FormatInt(goal_key.IntID(), 10) + "\""))
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
