package plugin

type Schedule struct {
	Frequency string `json:"frequency"`
	Interval  string `json:"interval"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time,omitempty"`
}
