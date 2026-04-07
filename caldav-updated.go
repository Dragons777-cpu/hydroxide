package caldav

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/hydroxide/protonmail"
)

var errNotFound = errors.New("caldav: not found")

var calendar = &caldav.Calendar{
	Path:            "/calendars/default",
	Name:            "ProtonMail",
	Description:     "ProtonMail calendar",
	MaxResourceSize: 1024 * 1024,
	SupportedComponentSet: []caldav.Component{
		caldav.ComponentEvent,
	},
}

// parseCalendarPath parses a calendar object path
func parseCalendarPath(p string) (string, error) {
	dirname, filename := path.Split(p)
	ext := path.Ext(filename)
	if dirname != "/calendars/default/" || ext != ".ics" {
		return "", errNotFound
	}
	return strings.TrimSuffix(filename, ext), nil
}

// formatCalendarPath formats a calendar object path
func formatCalendarPath(id string) string {
	return "/calendars/default/" + id + ".ics"
}

// toCalendarObject converts a ProtonMail calendar event to a CalDAV object
func (b *backend) toCalendarObject(event *protonmail.CalendarEvent, req *caldav.CalendarDataRequest) (*caldav.CalendarObject, error) {
	// TODO: decrypt and convert event to iCalendar format
	cal := ical.NewCalendar()
	cal.Version = "2.0"
	cal.ProductID = "-//ProtonMail AG//ProtonMail Calendar//EN"

	// TODO: properly convert ProtonMail event to iCalendar
	// For now, create a minimal event
	// In production, parse the event data and convert properly

	return &caldav.CalendarObject{
		Path:    formatCalendarPath(event.ID),
		ModTime: event.LastEditTime.Time(),
		ETag:    fmt.Sprintf("%x%x", event.LastEditTime, len(event.SharedEvents)),
		Data:    cal,
	}, nil
}

type backend struct {
	c           *protonmail.Client
	cache       map[string]*protonmail.CalendarEvent
	locker      sync.Mutex
	total       int
	privateKeys openpgp.EntityList
	calendarID  string
}

func (b *backend) CurrentUserPrincipal(ctx context.Context) (string, error) {
	return "/", nil
}

func (b *backend) CalendarHomeSetPath(ctx context.Context) (string, error) {
	return "/calendars", nil
}

func (b *backend) CreateCalendar(ctx context.Context, cal *caldav.Calendar) error {
	return webdav.NewHTTPError(http.StatusForbidden, errors.New("cannot create new calendar"))
}

func (b *backend) DeleteCalendar(ctx context.Context, path string) error {
	return webdav.NewHTTPError(http.StatusForbidden, errors.New("cannot delete calendar"))
}

func (b *backend) ListCalendars(ctx context.Context) ([]caldav.Calendar, error) {
	// Get calendar list from ProtonMail
	calendars, err := b.c.ListCalendars(0, 10)
	if err != nil {
		return nil, err
	}

	// Convert to CalDAV calendars
	result := make([]caldav.Calendar, 0, len(calendars))
	for _, cal := range calendars {
		result = append(result, caldav.Calendar{
			Path:        "/calendars/" + cal.ID,
			Name:        cal.Name,
			Description: cal.Description,
		})
	}

	if len(result) == 0 {
		return []caldav.Calendar{*calendar}, nil
	}

	return result, nil
}

func (b *backend) GetCalendar(ctx context.Context, path string) (*caldav.Calendar, error) {
	// For simplicity, return default calendar
	if path == calendar.Path || path == "/calendars/default" {
		return calendar, nil
	}
	return nil, webdav.NewHTTPError(http.StatusNotFound, errors.New("calendar not found"))
}

func (b *backend) cacheComplete() bool {
	b.locker.Lock()
	defer b.locker.Unlock()
	return b.total >= 0 && len(b.cache) == b.total
}

func (b *backend) getCache(id string) (*protonmail.CalendarEvent, bool) {
	b.locker.Lock()
	event, ok := b.cache[id]
	b.locker.Unlock()
	return event, ok
}

func (b *backend) putCache(event *protonmail.CalendarEvent) {
	b.locker.Lock()
	b.cache[event.ID] = event
	b.locker.Unlock()
}

func (b *backend) deleteCache(id string) {
	b.locker.Lock()
	delete(b.cache, id)
	b.locker.Unlock()
}

func (b *backend) GetCalendarObject(ctx context.Context, path string, req *caldav.CalendarDataRequest) (*caldav.CalendarObject, error) {
	id, err := parseCalendarPath(path)
	if err != nil {
		return nil, err
	}

	event, ok := b.getCache(id)
	if !ok {
		if b.cacheComplete() {
			return nil, errNotFound
		}

		// Fetch from ProtonMail API
		// Note: This requires knowing the calendarID
		// For now, return not found
		return nil, errNotFound
	}

	return b.toCalendarObject(event, req)
}

func (b *backend) ListCalendarObjects(ctx context.Context, path string, req *caldav.CalendarDataRequest) ([]caldav.CalendarObject, error) {
	if b.cacheComplete() {
		b.locker.Lock()
		defer b.locker.Unlock()

		cos := make([]caldav.CalendarObject, 0, len(b.cache))
		for _, event := range b.cache {
			co, err := b.toCalendarObject(event, req)
			if err != nil {
				return nil, err
			}
			cos = append(cos, *co)
		}

		return cos, nil
	}

	// Fetch from ProtonMail API
	// Use default calendar for now
	filter := &protonmail.CalendarEventFilter{
		Start:    time.Now().AddDate(-1, 0, 0).Unix(),
		End:      time.Now().AddDate(1, 0, 0).Unix(),
		Timezone: "UTC",
		Page:     0,
		PageSize: 100,
	}

	events, err := b.c.ListCalendarEvents(b.calendarID, filter)
	if err != nil {
		return nil, err
	}

	b.locker.Lock()
	b.total = len(events)
	b.locker.Unlock()

	cos := make([]caldav.CalendarObject, 0, len(events))
	for _, event := range events {
		b.putCache(event)
		co, err := b.toCalendarObject(event, req)
		if err != nil {
			return nil, err
		}
		cos = append(cos, *co)
	}

	return cos, nil
}

func (b *backend) QueryCalendarObjects(ctx context.Context, path string, query *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	req := caldav.CalendarDataRequest{AllProp: true}
	if query != nil {
		req = query.DataRequest
	}

	// Fetch all objects
	all, err := b.ListCalendarObjects(ctx, calendar.Path, &req)
	if err != nil {
		return nil, err
	}

	// TODO: implement proper filtering based on query
	return all, nil
}

func (b *backend) PutCalendarObject(ctx context.Context, path string, cal *ical.Calendar, opts *caldav.PutCalendarObjectOptions) (co *caldav.CalendarObject, error) {
	id, err := parseCalendarPath(path)
	if err != nil {
		return nil, err
	}

	// TODO: convert iCalendar to ProtonMail format and create/update event
	// For now, return not implemented
	return nil, webdav.NewHTTPError(http.StatusNotImplemented, errors.New("calendar event creation not yet implemented"))
}

func (b *backend) DeleteCalendarObject(ctx context.Context, path string) error {
	id, err := parseCalendarPath(path)
	if err != nil {
		return err
	}

	// TODO: implement delete via ProtonMail API
	return webdav.NewHTTPError(http.StatusNotImplemented, errors.New("calendar event deletion not yet implemented"))
}

func NewHandler(c *protonmail.Client, privateKeys openpgp.EntityList, calendarID string) http.Handler {
	if len(privateKeys) == 0 {
		panic("hydroxide/caldav: no private key available")
	}

	b := &backend{
		c:           c,
		cache:       make(map[string]*protonmail.CalendarEvent),
		total:       -1,
		privateKeys: privateKeys,
		calendarID:  calendarID,
	}

	return &caldav.Handler{Backend: b}
}
