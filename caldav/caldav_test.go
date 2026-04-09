package caldav

import (
	"testing"
	"time"
)

func TestFormatICS(t *testing.T) {
	cal := &Calendar{
		ID:          "cal1",
		Name:        "Test Calendar",
		Description: "A test calendar",
	}
	
	ev := &Event{
		ID:       "ev1",
		UID:      "test-uid-123",
		Summary:  "Test Event",
		Start:    time.Date(2026, 4, 10, 14, 0, 0, 0, time.UTC),
		End:      time.Date(2026, 4, 10, 15, 0, 0, 0, time.UTC),
		Location: "Test Location",
		AllDay:   false,
	}
	
	ics := formatICS(cal, ev)
	
	if !contains(ics, "BEGIN:VCALENDAR") {
		t.Error("Missing VCALENDAR start")
	}
	if !contains(ics, "SUMMARY:Test Event") {
		t.Error("Missing event summary")
	}
	if !contains(ics, "LOCATION:Test Location") {
		t.Error("Missing location")
	}
	if !contains(ics, "END:VCALENDAR") {
		t.Error("Missing VCALENDAR end")
	}
}

func TestFormatICSAllDay(t *testing.T) {
	cal := &Calendar{
		ID:   "cal1",
		Name: "Test",
	}
	
	ev := &Event{
		ID:      "ev1",
		UID:     "all-day-uid",
		Summary: "All Day Event",
		Start:   time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
		End:     time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC),
		AllDay:  true,
	}
	
	ics := formatICS(cal, ev)
	
	if !contains(ics, "DTSTART;VALUE=DATE") {
		t.Error("All-day event should have VALUE=DATE")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		(s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
