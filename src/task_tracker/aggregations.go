package aggregation

import (
	""
	"time"
	"sort"
)

type Aggregation struct {
	Recorded time.Time
	Success  bool
	Count    int
}

type byReported []task_tracker.Progress

// Implement sort.Interface
func (p byReported) Len() int { return len(p) }
func (p byReported) Swap(a int, b int) { p[a], p[b] = p[b], p[a] }
func (p byReported) Less(a int, b int) { return p[a].Reported.Before(p[b]) }

func aggregateFromTime(last_agg_time time.Time, reports []task_tracker.Progress) (*Aggregation, []task_tracker.Progress) {
	p := make([]task_tracker.Progress, 0)
	return nil, p
}

func GetAggregations(last_aggregation *Aggregation, period task_tracker.Period, frequency int, progress []task_tracker.Progress) {
	if len(progress) < 1 {
		// TODO: what is return value?
		return
	}
	sort.Sort(byReported(progress))
	last_agg_time := progress[0].Reported
	if last_aggregation != nil {
		last_agg_time = last_aggregation.Recorded
	}

	aggregations := make([]*Aggregation, 0)
	for {
		aggregation, remaining_reports := aggregateFromTime(last_agg_time, progress)
		if aggregation != nil {
			append(aggregations, aggregation)
		}
		if len(remaining_reports < 1) {
			break
		}
		last_agg_time = remaining_reports[0].Reported
	}



	//now := time.Now()
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
