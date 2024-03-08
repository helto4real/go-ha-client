package client

type HaResult struct {
	Id          int64       `json:"id"`
	MessageType string      `json:"type"`
	Success     bool        `json:"success"`
	Result      interface{} `json:"result"`
	Event       *HaEvent    `json:"event"`
	Error       *HaError    `json:"error"`
}

type HaError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type HaConnectMsg struct {
	MessageType string `json:"type"`
	HaVersion   string `json:"ha_version"`
}

type HaAuthMsg struct {
	AccessToken string `json:"access_token"`
	MessageType string `json:"type"`
}

type HaEvent struct {
	Data      EventData `json:"data"`
	EventType string    `json:"event_type"`
	TimeFired string    `json:"time_fired"`
}

type EventData struct {
	EntityId    string                 `json:"entity_id"`
	NewState    StateData              `json:"new_state"`
	OldState    StateData              `json:"old_state"`
	Domain      string                 `json:"domain"`
	Service     string                 `json:"service"`
	ServiceData map[string]interface{} `json:"service_data"`
}
type StateData struct {
	LastChanged string                 `json:"last_changed"`
	LastUpdated string                 `json:"last_updated"`
	State       interface{}            `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
}

type HaSubscribeEventsCommand struct {
	Id          int64  `json:"id"`
	MessageType string `json:"type"`
	EventType   string `json:"event_type"`
}
