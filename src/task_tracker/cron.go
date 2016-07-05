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

var defaultTasks = []Task{Task{"Walk"}, Task{"Run"}, Task{"Drink Water"}}

func resetTasksForUser(ctx appengine.Context, user_key *datastore.Key) ([]*datastore.Key, error) {
	task_keys, err:= datastore.NewQuery("Task").Ancestor(user_key).KeysOnly().GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}
	err = datastore.DeleteMulti(ctx, task_keys)
	if err != nil {
		return nil, err
	}
	task_keys_stub := []*datastore.Key{
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

func deleteProgressForGoal(wg *sync.WaitGroup, ctx appengine.Context, goal_key *datastore.Key) {
	defer wg.Done()
	progress_keys, err := datastore.NewQuery("Progress").Ancestor(goal_key).KeysOnly().GetAll(ctx, nil)
	if err != nil {
		ctx.Errorf("Failed to fetch progress for %v: %v", goal_key, err)
		return
	}
	err = datastore.DeleteMulti(ctx, progress_keys)
	if err != nil {
		ctx.Errorf("Failed to delete progress %v: %v", progress_keys, err)
		return
	}
}

func deleteAggregationsForGoal(wg *sync.WaitGroup, ctx appengine.Context, goal_key *datastore.Key) {
	defer wg.Done()
	progress_keys, err := datastore.NewQuery("Aggregation").Ancestor(goal_key).KeysOnly().GetAll(ctx, nil)
	if err != nil {
		ctx.Errorf("Failed to fetch aggregation for %v: %v", goal_key, err)
		return
	}
	err = datastore.DeleteMulti(ctx, progress_keys)
	if err != nil {
		ctx.Errorf("Failed to delete aggregation %v: %v", progress_keys, err)
		return
	}
}

func resetGoalsForUser(ctx appengine.Context, user_key *datastore.Key, task_keys []*datastore.Key) ([]*datastore.Key, error) {
	goal_keys, err := datastore.NewQuery("Goal").Ancestor(user_key).KeysOnly().GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}
	var pg sync.WaitGroup
	for _, goal_key := range goal_keys {
		pg.Add(1)
		go deleteProgressForGoal(&pg, ctx, goal_key)
		pg.Add(1)
		go deleteAggregationsForGoal(&pg, ctx, goal_key)
	}
	pg.Wait()
	err = datastore.DeleteMulti(ctx, goal_keys)
	if err != nil {
		return nil, err
	}
	goal_keys_stub := []*datastore.Key{
		datastore.NewIncompleteKey(ctx, "Goal", user_key),
		datastore.NewIncompleteKey(ctx, "Goal", user_key),
		datastore.NewIncompleteKey(ctx, "Goal", user_key),
	}
	goals := []Goal{
		Goal{
			Task: task_keys[0],
			Period: Day,
			Frequency: 2,
		},
		Goal{
			Task: task_keys[1],
			Period: Week,
			Frequency: 4,
		},
		Goal{
			Task: task_keys[2],
			Period: Day,
			Frequency: 5,
		},
	}
	new_goal_keys, err := datastore.PutMulti(ctx, goal_keys_stub, goals)
	if err != nil {
		return nil, err
	}
	return new_goal_keys, nil
}

func progress(t time.Time) *Progress {
	return &Progress{
		Reported: t,
		Aggregated: false,
	}
}

func progress_key(ctx appengine.Context, goal_key *datastore.Key) *datastore.Key {
	return datastore.NewIncompleteKey(ctx, "Progress", goal_key)
}

func resetProgressForUser(ctx appengine.Context, user_key *datastore.Key, goal_keys []*datastore.Key) error {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	two_days_ago := yesterday.Add(-24 * time.Hour)

	walk_progress := []*Progress{
		progress(two_days_ago), progress(two_days_ago),
		progress(yesterday),
		progress(now),
	}
	walk_progress_key_stubs := []*datastore.Key{
		progress_key(ctx, goal_keys[0]),
		progress_key(ctx, goal_keys[0]),
		progress_key(ctx, goal_keys[0]),
		progress_key(ctx, goal_keys[0]),
	}
	_, err := datastore.PutMulti(ctx, walk_progress_key_stubs, walk_progress)
	if err != nil {
		return err
	}

	run_progress := []*Progress{progress(two_days_ago), progress(yesterday)}
	run_progress_key_stubs := []*datastore.Key{
		progress_key(ctx, goal_keys[1]),
		progress_key(ctx, goal_keys[1]),
	}
	_, err = datastore.PutMulti(ctx, run_progress_key_stubs, run_progress)
	if err != nil {
		return err
	}

	water_progress := []*Progress{
		progress(two_days_ago), progress(two_days_ago), progress(two_days_ago), progress(two_days_ago),
		progress(two_days_ago), progress(two_days_ago), progress(two_days_ago),
		progress(yesterday), progress(yesterday), progress(yesterday), progress(yesterday),
		progress(yesterday), progress(yesterday), progress(yesterday),
		progress(now), progress(now), progress(now),
	}
	water_progress_key_stubs := []*datastore.Key{
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
		progress_key(ctx, goal_keys[2]),
	}
	_, err = datastore.PutMulti(ctx, water_progress_key_stubs, water_progress)
	if err != nil {
		return err
	}
	return nil
}

func resetUser(wg *sync.WaitGroup, ctx appengine.Context, user_key *datastore.Key) {
	defer wg.Done()
	task_keys, err := resetTasksForUser(ctx, user_key)
	if err != nil {
		ctx.Errorf("Task reset failed for %v: %v", user_key, err)
		return
	}
	goal_keys, err := resetGoalsForUser(ctx, user_key, task_keys)
	if err != nil {
		ctx.Errorf("Goal reset failed for %v: %v", user_key, err)
		return
	}
	err = resetProgressForUser(ctx, user_key, goal_keys)
	if err != nil {
		ctx.Errorf("Failed to reset progress for %v: %v", user_key, err)
	}
}

func resetTestData(w http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	var wg sync.WaitGroup
	users_query := datastore.NewQuery("User").KeysOnly()
	users_iter := users_query.Run(ctx)
	for {
		user_key, err := users_iter.Next(nil)
		if err == datastore.Done {
			break
		} else if err != nil {
			ctx.Errorf("Failed to get key: %v, stopping iteration", err)
			break
		}
		wg.Add(1)
		go resetUser(&wg, ctx, user_key)
	}
	wg.Wait()
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
