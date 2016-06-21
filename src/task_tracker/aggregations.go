package task_tracker

import (
	"appengine/datastore"
	"sort"
	"time"
	"appengine"
)

type ProgressEntity struct {
	Key    *datastore.Key
	Report *Progress
}

type Aggregation struct {
	Recorded time.Time
	Success  bool
	Count    int
}

type AggregationResults struct {
	Aggregations         []*Aggregation
	ProgressesAggregated []*datastore.Key
}

type byReported []*ProgressEntity

// Implement sort.Interface
func (p byReported) Len() int               { return len(p) }
func (p byReported) Swap(a int, b int)      { p[a], p[b] = p[b], p[a] }
func (p byReported) Less(a int, b int) bool { return p[a].Report.Reported.Before(p[b].Report.Reported) }

func (p Period) hasElapsedSince(from time.Time, to time.Time) bool {
	switch p {
	case Day:
		return from.Add(24 * time.Hour).Before(to)
	}
	return false
}

func (p Period) toPeriodStart(t time.Time) time.Time {
	switch p {
	case Day:
		return t.Truncate(24 * time.Hour)
	}
	return t
}

func aggregateFromTime(period_start time.Time, now time.Time, reports []*ProgressEntity) (*Aggregation, []*datastore.Key, []*ProgressEntity) {
	found_reports := 0
	first_in_next_period := -1
	for i, report_entity := range reports {

	}
	p := make([]*ProgressEntity, 0)
	return nil, nil, p
}

func GetAggregations(ctx appengine.Context, last_aggregation *Aggregation, period Period, frequency int, progress []*ProgressEntity) *AggregationResults {
	if len(progress) < 1 {
		// TODO: what is return value?
		return nil
	}
	sort.Sort(byReported(progress))
	last_agg_time := progress[0].Report.Reported
	if last_aggregation != nil {
		last_agg_time = last_aggregation.Recorded
	}
	now := time.Now()
	aggregations := make([]*Aggregation, 0)
	all_processed_keys := make([]*datastore.Key, 0)
	for {
		last_agg_time = period.toPeriodStart(last_agg_time)
		from := last_agg_time
		to := now
		elapsed := period.hasElapsedSince(from, to)
		ctx.Debugf("Period elapsed from %v to %v? %v", from, to, elapsed)
		if !period.hasElapsedSince(last_agg_time, now) {
			break
		}
		aggregation, processed_keys, remaining_reports := aggregateFromTime(last_agg_time, now, progress)
		if aggregation != nil {
			aggregations = append(aggregations, aggregation)
			all_processed_keys = append(all_processed_keys, processed_keys...)
		}
		if len(remaining_reports) < 1 {
			break
		}
		last_agg_time = remaining_reports[0].Report.Reported
	}

	return &AggregationResults{aggregations, all_processed_keys}
	// TODO: implement this
	//
	// Find last aggregation time
	// Determine if period has elapsed
	// if not, return
	// sort progress
	// take progresses until end of next period
	// bucket them
	// repeat
}
