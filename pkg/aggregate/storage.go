package aggregate

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/golang/glog"

	"github.com/myntra/aggo/pkg/event"
)

type storage struct {
	mu          sync.Mutex
	m           map[string]*event.RuleBucket // [ruleID]
	flusherChan chan string
}

func isMatch(eventType, pattern string) bool {
	return eventType == pattern
}

func (s *storage) stash(event *event.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: dedup events in a window
	// TODO: efficient regex matching for rule bucket
	// TODO: sliding wait window
	// TODO: frequency count of event

	for ruleID, ruleBucket := range s.m {
		for _, eventTypePattern := range ruleBucket.Rule.EventTypes {
			// add event to all matching rule buckets
			if isMatch(event.EventType, eventTypePattern) {
				if len(s.m[ruleID].Bucket) == 0 {
					// start a flusher for this rule
					go func() {
						time.Sleep(time.Millisecond * time.Duration(ruleBucket.Rule.WaitWindow))
						rb := s.getRule(ruleID)
						if rb == nil {
							glog.Errorf("unexpected err ruleID %v not found", ruleID)
							return
						}
						err := rb.Post()
						if err != nil {
							b, err2 := json.Marshal(rb)
							glog.Errorf("post rule bucket failed. dropping it!! %v %v %v", err, string(b), err2)
						}
						s.flusherChan <- ruleID
					}()
				}
				// dedup, reschedule flusher(sliding wait window), frequency count
				s.m[ruleID].Bucket = append(s.m[ruleID].Bucket, event)
			}
		}
	}
}

func (s *storage) getRule(ruleID string) *event.RuleBucket {
	s.mu.Lock()
	defer s.mu.Unlock()
	var rb *event.RuleBucket
	var ok bool
	if rb, ok = s.m[ruleID]; !ok {
		return nil
	}
	return rb
}

func (s *storage) flushRule(ruleID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[ruleID].Bucket = nil
}

func (s *storage) addRule(rule *event.Rule) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[rule.ID]; ok {
		// rule id already exists
		return false
	}

	ruleBucket := &event.RuleBucket{
		Rule: rule,
	}

	s.m[rule.ID] = ruleBucket

	return true
}

func (s *storage) removeRule(ruleID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[ruleID]; !ok {
		// rule id does not exist
		return false
	}

	delete(s.m, ruleID)

	return true
}

func (s *storage) getRules() []*event.Rule {
	s.mu.Lock()
	defer s.mu.Unlock()
	var rules []*event.Rule
	for _, v := range s.m {
		rules = append(rules, v.Rule)
	}
	return rules
}

func (s *storage) clone() map[string]*event.RuleBucket {
	s.mu.Lock()
	defer s.mu.Unlock()
	clone := make(map[string]*event.RuleBucket)
	for k, v := range s.m {
		clone[k] = v
	}
	return clone
}

func (s *storage) restore(m map[string]*event.RuleBucket) {
	s.m = m
}
