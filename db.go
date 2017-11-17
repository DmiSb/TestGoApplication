package main

import (
	"time"
)

type StatStorage struct {
	Country string
	Os      string
	App     string
	Pos     int
	Last    time.Time
	Count   int
}
