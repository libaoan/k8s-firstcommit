package util

import (
	"encoding/json"
	"log"
	"time"
)

// Simply catches a crash and logs an error. Meant to be called via defer.
func HandleCrash() {
	r := recover()
	if r != nil {
		log.Printf("Recovery from panic: %#v", r)
	}
}

// Loops forever running f every d.  Catches any panics, and keeps going.
func Forever(f func(), period time.Duration) {
	for {
		func() {
			defer HandleCrash()
			f()
		}()
		time.Sleep(period)
	}
}

// Returns o marshalled as a JSON string, ignoring any errors.
func MakeJSONString(o interface{}) string {
	data, _ := json.Marshal(o)
	return string(data)
}
