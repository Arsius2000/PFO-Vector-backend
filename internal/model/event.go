package model

type EventResponse struct {
	ID        int32  `json:"id"`
	EventDate string `json:"event_date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Title     string `json:"title"`
	Audience  string `json:"audience"`
	Weight    int32  `json:"weight"`
	CreatedBy int32  `json:"created_by"`
}

type EventsListResponse struct {
	Events     []EventResponse `json:"events"`
	Pagination Pagination      `json:"pagination"`
}
