package main

import "time"

func createdAt(t time.Time) float64 {
	time := checkTimestamp(t)
	return time
}
func checkTimestamp(t time.Time) float64 {
	loc, _ := time.LoadLocation("UTC")
	expiresAt := time.Now().In(loc).Add(0 * time.Hour)

	return expiresAt.Sub(t).Minutes()
}
