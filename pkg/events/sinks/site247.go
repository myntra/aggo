package sinks

import (
	"fmt"
	"time"

	"github.com/myntra/cortex/pkg/events"
	"github.com/golang/glog"
	"github.com/satori/go.uuid"
)

type Site247Alert struct {
	MonitorName          string `json:"MONITORNAME,omitempty"`
	MonitorGroupName     string `json:"MONITOR_GROUPNAME,omitempty"`
	SearchPollFrequency  int    `json:"SEARCH POLLFREQUENCY,omitempty"`
	MonitorID            int    `json:"MONITOR_ID,omitempty"`
	FailedLocations      string `json:"FAILED_LOCATIONS,omitempty"`
	MonitorURL           string `json:"MONITORURL,omitempty"`
	IncidentTimeISO      string `json:"INCIDENT_TIME_ISO,omitempty"`
	MonitorType          string `json:"MONITORTYPE,omitempty"`
	Status               string `json:"STATUS,omitempty"`
	Timezone             string `json:"TIMEZONE,omitempty"`
	IncidentTime         string `json:"INCIDENT_TIME,omitempty"`
	IncidentReason       string `json:"INCIDENT_REASON,omitempty"`
	OutageTimeUnixFormat int    `json:"OUTAGE_TIME_UNIX_FORMAT,omitempty"`
	RCALink              string `json:"RCA_LINK,omitempty"`
}


// EventFromSite247 converts alerts sent from site24x7 into cloud events
func EventFromSite247(alert Site247Alert) *events.Event {
	event := events.Event{
		Source:             "site247",
		Data:               alert,
		ContentType:        "application/json",
		EventTypeVersion:   "1.0",
		CloudEventsVersion: "0.1",
		SchemaURL:          "",
		EventID:            generateUUID().String(),
		EventTime:          time.Now(),
		EventType:          fmt.Sprintf("site247.%s.%s", alert.MonitorGroupName, alert.MonitorName),
	}
	return &event
}

func generateUUID() uuid.UUID {
	uid, err := uuid.NewV4()
	if err != nil {
		glog.Infof("Error in creating new UUID for event sink")
		return uuid.UUID{}
	}
	return uid
}
