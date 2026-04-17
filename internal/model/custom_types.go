package model

import (
	"strings"
	"time"
)

// Для формата "02.01.2006" (или 02-01-2006)
type CustomDate time.Time

func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" || s == "null" {
		return nil
	}
	// Укажи здесь разделитель, который будет присылать фронт (. или -)
	t, err := time.Parse("02.01.2006", s) 
	if err != nil {
		return err
	}
	*cd = CustomDate(t)
	return nil
}

// Для формата "15:04"
type CustomTime time.Time

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.Parse("15:04", s)
	if err != nil {
		return err
	}
	*ct = CustomTime(t)
	return nil
}
