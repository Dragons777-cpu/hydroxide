// Package caldav provides CalDAV calendar sync support for hydroxide.
// It exposes calendars and events through a local CalDAV server.
package caldav

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/hydroxide/protonmail"
)

// Handler handles CalDAV requests.
type Handler struct {
	client *protonmail.Client
	uid    string
}

// NewHandler creates a new CalDAV handler.
func NewHandler(client *protonmail.Client, uid string) *Handler {
	return &Handler{
		client: client,
		uid:    uid,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: implement proper CalDAV handling
	// For now, just return a basic response
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<?xml version="1.0" encoding="utf-8"?>
<multistatus xmlns="DAV:">
</multistatus>`)
}

// Calendar represents a ProtonMail calendar.
type Calendar struct {
	ID          string
	Name        string
	Description string
	Color       string
}

// Event represents a calendar event.
type Event struct {
	ID        string
	UID       string
	Summary   string
	Start     time.Time
	End       time.Time
	Location  string
	AllDay    bool
	RawICS    string
}

// ListCalendars returns all calendars for the user.
func (h *Handler) ListCalendars() ([]*Calendar, error) {
	calendars, err := h.client.GetCalendars()
	if err != nil {
		return nil, err
	}

	var res []*Calendar
	for _, cal := range calendars {
		res = append(res, &Calendar{
			ID:          cal.ID,
			Name:        cal.Name,
			Description: cal.Description,
			Color:       cal.Color,
		})
	}
	return res, nil
}

// GetEvents returns events for a calendar within a time range.
func (h *Handler) GetEvents(calendarID string, start, end time.Time) ([]*Event, error) {
	events, err := h.client.GetCalendarEvents(calendarID, start, end)
	if err != nil {
		return nil, err
	}

	var res []*Event
	for _, ev := range events {
		res = append(res, &Event{
			ID:       ev.ID,
			UID:      ev.UID,
			Summary:  ev.Summary,
			Start:    ev.Start,
			End:      ev.End,
			Location: ev.Location,
			AllDay:   ev.AllDay,
		})
	}
	return res, nil
}

// GetEventICS returns an event in iCalendar format.
func (h *Handler) GetEventICS(calendarID, eventID string) (string, error) {
	event, err := h.client.GetCalendarEvent(calendarID, eventID)
	if err != nil {
		return "", err
	}
	return event.ICS, nil
}

// formatICS formats an event as iCalendar.
func formatICS(cal *Calendar, ev *Event) string {
	var buf bytes.Buffer
	
	buf.WriteString("BEGIN:VCALENDAR\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("PRODID:-//hydroxide//CalDAV//EN\r\n")
	buf.WriteString("CALSCALE:GREGORIAN\r\n")
	buf.WriteString("X-WR-CALNAME:" + cal.Name + "\r\n")
	
	buf.WriteString("BEGIN:VEVENT\r\n")
	buf.WriteString("UID:" + ev.UID + "\r\n")
	buf.WriteString("DTSTAMP:" + time.Now().Format("20060102T150405Z") + "\r\n")
	
	if ev.AllDay {
		buf.WriteString("DTSTART;VALUE=DATE:" + ev.Start.Format("20060102") + "\r\n")
		buf.WriteString("DTEND;VALUE=DATE:" + ev.End.Format("20060102") + "\r\n")
	} else {
		buf.WriteString("DTSTART:" + ev.Start.Format("20060102T150405Z") + "\r\n")
		buf.WriteString("DTEND:" + ev.End.Format("20060102T150405Z") + "\r\n")
	}
	
	buf.WriteString("SUMMARY:" + ev.Summary + "\r\n")
	if ev.Location != "" {
		buf.WriteString("LOCATION:" + ev.Location + "\r\n")
	}
	
	buf.WriteString("END:VEVENT\r\n")
	buf.WriteString("END:VCALENDAR\r\n")
	
	return buf.String()
}

// Backend implements caldav.Backend interface.
type Backend struct {
	handler *Handler
}

// NewBackend creates a new CalDAV backend.
func NewBackend(h *Handler) *Backend {
	return &Backend{handler: h}
}

// CurrentUserPrincipal returns the current user principal.
func (b *Backend) CurrentUserPrincipal(ctx *caldav.Context) (string, error) {
	return "/user/" + b.handler.uid, nil
}

// CalendarSet returns calendar home set.
func (b *Backend) CalendarSet(ctx *caldav.Context) (string, error) {
	return "/user/" + b.handler.uid + "/calendars/", nil
}

// Calendar returns a calendar by ID.
func (b *Backend) Calendar(id string) (*caldav.Calendar, error) {
	calendars, err := b.handler.ListCalendars()
	if err != nil {
		return nil, err
	}
	
	for _, cal := range calendars {
		if cal.ID == id {
			return &caldav.Calendar{
				Path:        "/calendars/" + cal.ID,
				Name:        cal.Name,
				Description: cal.Description,
			}, nil
		}
	}
	return nil, caldav.ErrNotFound
}

// ListCalendars returns all calendars.
func (b *Backend) ListCalendars(ctx *caldav.Context) ([]caldav.Calendar, error) {
	calendars, err := b.handler.ListCalendars()
	if err != nil {
		return nil, err
	}
	
	var res []caldav.Calendar
	for _, cal := range calendars {
		res = append(res, caldav.Calendar{
			Path:        "/calendars/" + cal.ID,
			Name:        cal.Name,
			Description: cal.Description,
		})
	}
	return res, nil
}

// GetCalendarObject returns an event.
func (b *Backend) GetCalendarObject(path string) (*caldav.CalendarObject, error) {
	// Parse path: /calendars/{calendarID}/{eventID}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		return nil, caldav.ErrNotFound
	}
	
	calendarID := parts[0]
	eventID := parts[1]
	
	ics, err := b.handler.GetEventICS(calendarID, eventID)
	if err != nil {
		return nil, err
	}
	
	return &caldav.CalendarObject{
		Path: path,
		Data: io.NopCloser(strings.NewReader(ics)),
	}, nil
}

// ListCalendarObjects returns all events in a calendar.
func (b *Backend) ListCalendarObjects(calendarPath string) ([]caldav.CalendarObject, error) {
	parts := strings.Split(strings.Trim(calendarPath, "/"), "/")
	if len(parts) == 0 {
		return nil, caldav.ErrNotFound
	}
	
	calendarID := parts[len(parts)-1]
	
	// Get events for the last year
	start := time.Now().AddDate(-1, 0, 0)
	end := time.Now().AddDate(1, 0, 0)
	
	events, err := b.handler.GetEvents(calendarID, start, end)
	if err != nil {
		return nil, err
	}
	
	var res []caldav.CalendarObject
	for _, ev := range events {
		calendars, _ := b.handler.ListCalendars()
		var cal *Calendar
		for _, c := range calendars {
			if c.ID == calendarID {
				cal = c
				break
			}
		}
		if cal == nil {
			continue
		}
		
		ics := formatICS(cal, ev)
		res = append(res, caldav.CalendarObject{
			Path: calendarPath + "/" + ev.ID,
			Data: io.NopCloser(strings.NewReader(ics)),
			ModTime: ev.Start,
		})
	}
	return res, nil
}

// PutCalendarObject creates or updates an event.
// TODO: implement write support
func (b *Backend) PutCalendarObject(path string, data io.ReadCloser) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// DeleteCalendarObject deletes an event.
// TODO: implement delete support
func (b *Backend) DeleteCalendarObject(path string) error {
	return fmt.Errorf("not implemented")
}
