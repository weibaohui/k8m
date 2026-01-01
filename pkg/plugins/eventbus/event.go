package eventbus

type EventType string

const (
	EventLeaderElected EventType = "leader.elected"
	EventLeaderLost    EventType = "leader.lost"
)

type Event struct {
	Type EventType
	Data any
}
