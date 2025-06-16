package webhooksender

import (
	"time"
)

// InspectionCheckEvent is the event to be pushed.
type InspectionCheckEvent struct {
	ID          uint      `json:"id"`
	RecordID    uint      `json:"record_id"`
	EventStatus string    `json:"event_status"`
	EventMsg    string    `json:"event_msg"`
	Extra       string    `json:"extra"`
	ScriptName  string    `json:"script_name"`
	Kind        string    `json:"kind"`
	CheckDesc   string    `json:"check_desc"`
	Cluster     string    `json:"cluster"`
	Namespace   string    `json:"namespace"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ScheduleID  *uint     `json:"schedule_id"`
}
