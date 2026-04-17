package model

import (
	"time"

	"pfo-vector/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

func pgTimeToTime(t pgtype.Time) time.Time {
	micro := t.Microseconds
	h := int(micro / int64(time.Hour/time.Microsecond))
	micro %= int64(time.Hour / time.Microsecond)
	m := int(micro / int64(time.Minute/time.Microsecond))
	micro %= int64(time.Minute / time.Microsecond)
	s := int(micro / int64(time.Second/time.Microsecond))
	micro %= int64(time.Second / time.Microsecond)
	ns := int(micro * int64(time.Microsecond))

	return time.Date(1, 1, 1, h, m, s, ns, time.UTC)
}

func MapEventFromRepo(e repository.Event) EventResponse {
	eventDate := time.Time{}
	if e.EventDate.Valid {
		eventDate = e.EventDate.Time
	}

	startTime := time.Time{}
	if e.StartTime.Valid {
		startTime = pgTimeToTime(e.StartTime)
	}

	endTime := time.Time{}
	if e.EndTime.Valid {
		endTime = pgTimeToTime(e.EndTime)
	}

	title := ""
	if e.Title.Valid {
		title = e.Title.String
	}

	audience := ""
	if e.Audience.Valid {
		audience = e.Audience.String
	}

	weight := int32(0)
	if e.Weight.Valid {
		weight = e.Weight.Int32
	}

	return EventResponse{
		ID:        e.ID,
		EventDate: eventDate,
		StartTime: startTime,
		EndTime:   endTime,
		Title:     title,
		Audience:  audience,
		Weight:    weight,
		CreatedBy: e.CreatedBy,
	}
}

func MapEventsFromRepo(items []repository.Event) []EventResponse {
	out := make([]EventResponse, 0, len(items))
	for _, e := range items {
		out = append(out, MapEventFromRepo(e))
	}
	return out
}
