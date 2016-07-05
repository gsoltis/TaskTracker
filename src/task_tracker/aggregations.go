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
	Recorded time.Time	`json:"-"`
	Completed time.Time
	Success  bool
	Count    int
}

type AggregationResults struct {
	Aggregations         []*Aggregation
	ProgressesAggregated []*ProgressEntity
}

func (results *AggregationResults) Record(ctx appengine.Context, goal_key *datastore.Key) error {
	err := datastore.RunInTransaction(ctx, func(ctx appengine.Context) error {
		keys := make([]*datastore.Key, 0, len(results.Aggregations))
		for i := 0; i < len(results.Aggregations); i++ {
			keys = append(keys, datastore.NewIncompleteKey(ctx, "Aggregation", goal_key))
		}
		_, err := datastore.PutMulti(ctx, keys, results.Aggregations)
		if err != nil {
			return err
		}
		progress_keys := make([]*datastore.Key, 0, len(results.ProgressesAggregated))
		progresses := make([]*Progress, 0, len(results.ProgressesAggregated))
		for _, pe := range results.ProgressesAggregated {
			progress_keys = append(progress_keys, pe.Key)
			pe.Report.Aggregated = true
			progresses = append(progresses, pe.Report)
		}
		_, err = datastore.PutMulti(ctx, progress_keys, progresses)
		if err != nil {
			return err
		}
		return nil
	}, nil)
	return err
}

type byReported []*ProgressEntity

// Implement sort.Interface
func (p byReported) Len() int               { return len(p) }
func (p byReported) Swap(a int, b int)      { p[a], p[b] = p[b], p[a] }
func (p byReported) Less(a int, b int) bool { return p[a].Report.Reported.Before(p[b].Report.Reported) }

func (p Period) hasElapsedSince(from time.Time, to time.Time) bool {
	end := p.addPeriod(from)
	return to.Equal(end) || to.After(end)
}

func (p Period) toPeriodStart(t time.Time) time.Time {
	switch p {
	case Day:
		return t.Truncate(24 * time.Hour)
	case Week:
		return t.Truncate(7 * 24 * time.Hour)
	}
	return t
}

func (p Period) addPeriod(from time.Time) time.Time {
	switch p {
	case Day:
		return from.Add(24 * time.Hour)
	case Week:
		return from.Add(7 * 24 * time.Hour)
	}
	return from
}

func aggregateFromTime(ctx appengine.Context, period_start time.Time, end_time time.Time, reports []*ProgressEntity) (*Aggregation, []*ProgressEntity, []*ProgressEntity) {
	found_reports := 0
	first_in_next_period := -1
	in_range := func(t time.Time) bool {
		return t.Equal(period_start) || (t.After(period_start) && t.Before(end_time))
	}
	processed := make([]*ProgressEntity, 0)
	for i, report_entity := range reports {
		if in_range(report_entity.Report.Reported) {
			found_reports += 1
			processed = append(processed, report_entity)
		} else {
			ctx.Debugf("Out of range: %v (%v %v)", report_entity.Report.Reported, period_start, end_time)
			first_in_next_period = i
			break
		}
	}

	if found_reports > 0 {
		agg := &Aggregation{
			Count: found_reports,
			Completed: end_time,
		}
		var remaining []*ProgressEntity
		if first_in_next_period == -1 {
			ctx.Debugf("Returning empty")
			remaining = make([]*ProgressEntity, 0)
		} else {
			remaining = reports[first_in_next_period:]
			ctx.Debugf("Returning from %v (%v)", first_in_next_period, len(remaining))
		}
		return agg, processed, remaining
	} else {
		ctx.Debugf("Found nothing between %v and %v. First: %v",
			period_start, end_time, reports[0].Report.Reported)
		return nil, nil, reports
	}
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
	aggregations := make([]*Aggregation, 0)
	aggregated_reports := make([]*ProgressEntity, 0)
	now := time.Now()
	ctx.Debugf("Starting with %v", last_agg_time)
	ctx.Debugf("Now? %v", now)
	count := 0
	for {
		last_agg_time = period.toPeriodStart(last_agg_time)
		end_time := period.addPeriod(last_agg_time)
		from := last_agg_time
		to := end_time
		elapsed := period.hasElapsedSince(from, to)
		ctx.Debugf("Period elapsed from %v to %v? %v (%v)", from, to, elapsed, count)
		if !period.hasElapsedSince(last_agg_time, now) {
			break
		}
		aggregation, processed_reports, remaining_reports := aggregateFromTime(ctx, last_agg_time, end_time, progress)
		if aggregation != nil {
			aggregation.Recorded = now
			aggregation.Success = aggregation.Count >= frequency
			aggregations = append(aggregations, aggregation)
			aggregated_reports = append(aggregated_reports, processed_reports...)
		}
		if len(remaining_reports) < 1 {
			break
		} else {
			ctx.Debugf("Remianing reports: %v", len(remaining_reports))
		}
		ctx.Debugf("Next remaining report: %v", remaining_reports[0].Report)
		last_agg_time = remaining_reports[0].Report.Reported
		ctx.Debugf("Last agg time? %v", last_agg_time)
		progress = remaining_reports
		count = count + 1
		//if count == 3 {
		//	break
		//}
	}
	results := &AggregationResults{aggregations, aggregated_reports}
	ctx.Debugf("Results: %v", results)
	return results
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
