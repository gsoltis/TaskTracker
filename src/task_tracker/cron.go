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
	GetAggregations(ctx, last_aggregation, goal.Period, goal.Frequency, progress)
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
	//ctx.Debugf("Request: %v", req)
	//ctx.Debugf("Cookies: %v", req.Cookies())
	fmt.Fprintf(w, "Cronnin'")
}
