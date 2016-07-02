package task_tracker

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
	"time"
)

func AddCronRoutes(r *mux.Router) {
	r.HandleFunc("/aggregate", aggregateProgress)
	r.HandleFunc("/resetTestData", resetTestData)
}

var defaultTasks = [3]Task{Task{"Walk"}, Task{"Run"}, Task{"Drink Water"}}

func resetTasksForUser(ctx appengine.Context, user_key *datastore.Key) ([]*datastore.Key, error) {
	task_keys, err:= datastore.NewQuery("Task").Ancestor(user_key).KeysOnly().GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}
	datastore.DeleteMulti(ctx, task_keys)
	task_keys_stub := [3]*datastore.Key{
		datastore.NewIncompleteKey(ctx, "Task", user_key),
		datastore.NewIncompleteKey(ctx, "Task", user_key),
		datastore.NewIncompleteKey(ctx, "Task", user_key),
	}
	new_task_keys, err := datastore.PutMulti(ctx, task_keys_stub, defaultTasks)
	if err != nil {
		return nil, err
	}
	return new_task_keys, nil
}

func resetGoalsForUser(ctx appengine.Context, user_key *datastore.Key, task_keys []*datastore.Key) ([]*datastore.Key, error) {

}

func resetUser(wg sync.WaitGroup, ctx appengine.Context, user_key *datastore.Key) {
	//task_keys, err := resetTasksForUser(user_key)
	//goal_keys, err := resetGoalsForUser(user_key, task_keys)
	//resetProgressForUser(user_key, goal_keys)
	wg.Done()
}

func resetTestData(w http.ResponseWriter, req *http.Request) {

}

func aggregateGoal(wg *sync.WaitGroup, ctx appengine.Context, goal_key_string string, progress []*ProgressEntity) {
	defer wg.Done()
	goal_key, err := datastore.DecodeKey(goal_key_string)
	if err != nil {
		ctx.Errorf("Failed to deserialize goal key: %v", err)
		return
	}
	goal := &Goal{}
	err = datastore.Get(ctx, goal_key, goal)
	if err != nil {
		ctx.Errorf("Failed to fetch goal: %v", err)
		return
	}
	ctx.Infof("Aggregating goal %v", goal_key)
	aggregations := make([]Aggregation, 0, 1)
	_, err = datastore.NewQuery("Aggregation").Ancestor(goal_key).Order("-Recorded").Limit(1).GetAll(ctx, &aggregations)
	if err != nil {
		ctx.Errorf("Failed to fetch aggregations")
		return
	}
	ctx.Infof("Fetched aggregation: %v", aggregations)
	var last_aggregation *Aggregation = nil
	if len(aggregations) > 0 {
		last_aggregation = &aggregations[0]
	}
	agg_results := GetAggregations(ctx, last_aggregation, goal.Period, goal.Frequency, progress)
	err = agg_results.Record(ctx, goal_key)
	if err != nil {
		ctx.Errorf("Failed to record aggregations")
		return
	}
}

func aggregateUser(wg *sync.WaitGroup, ctx appengine.Context, user_key *datastore.Key, aggregation_time *time.Time) {
	defer wg.Done()
	ctx.Debugf("Aggregating %v from ", user_key, aggregation_time)
	progresses := make([]Progress, 0)
	keys, err := datastore.NewQuery("Progress").Ancestor(user_key).GetAll(ctx, &progresses)
	if err != nil {
		ctx.Errorf("Failed, %v", err)
		return
	}
	ctx.Debugf("Got progress: %v", progresses)
	goal_keys := make(map[string][]*ProgressEntity)
	for i, key := range keys {
		goal_key := key.Parent().Encode()
		goal_progress, found := goal_keys[goal_key]
		if !found {
			goal_progress = make([]*ProgressEntity, 0)
		}
		goal_progress = append(goal_progress, &ProgressEntity{key, &progresses[i]})
		goal_keys[goal_key] = goal_progress
	}

	ctx.Infof("Got progress: %v", goal_keys)
	var goal_wg sync.WaitGroup
	for goal_key_string, goal_progress := range goal_keys {
		goal_wg.Add(1)
		go aggregateGoal(&goal_wg, ctx, goal_key_string, goal_progress)
	}
	goal_wg.Wait()
}

func aggregateProgress(w http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	var wg sync.WaitGroup

	users_query := datastore.NewQuery("User").Project("LastAggregation")
	users_iter := users_query.Run(ctx)
	for {
		var aggregation_time struct{ LastAggregation time.Time }
		user_key, err := users_iter.Next(&aggregation_time)
		if err == datastore.Done {
			break
		} else if err != nil {
			ctx.Errorf("Failed to get key: %v, stopping iteration", err)
			break
		}
		ctx.Debugf("adding 1")
		wg.Add(1)
		go aggregateUser(&wg, ctx, user_key, &aggregation_time.LastAggregation)
	}
	wg.Wait()
	fmt.Fprintf(w, "Aggregated")
}
