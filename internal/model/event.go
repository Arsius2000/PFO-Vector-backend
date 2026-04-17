package model

import "time"

type EventResponse struct {
	ID        int32 `json:"id"`
	EventDate time.Time `json:"event_date"`
	StartTime time.Time `json:"start_time"`
	EndTime time.Time `json:"end_time"`
	Title string `json:"title"`
	Audience string `json:"audience"`
	Weight int32 `json:"weight"`
	CreatedBy int32 `json:"created_by"`
	 
}

type EventsListResponse struct {
    Events []EventResponse `json:"events"`
    Pagination Pagination `json:"pagination"`
}