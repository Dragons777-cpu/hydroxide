package caldav

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/hydroxide/protonmail"
)

// TODO: use a HTTP error
var errNotFound = errors.New("caldav: not found")

var calendar = &caldav.Calendar{
	Path:            "/calendars/default",
	Name:            "ProtonMail",
	Description:     "ProtonMail calendar",
	MaxResourceSize: 1024 * 1024, // 1MB
	SupportedComponentSet: []caldav.Component{
		caldav.ComponentEvent,
	},
}

// formatCalendar formats an iCalendar for ProtonMail
func formatCalendar(cal *ical.Calendar, privateKey *openpgp.Entity) (*protonmail.CalendarImport, error) {
	var b bytes.Buffer
	if err := ical.NewEncoder(&b).Encode(cal); err != nil {
		return nil, err
	}

	// Encrypt the calendar data
	to := []*openpgp.Entity{privateKey}
	encrypted, err := protonmail.NewEncryptedCalendarCard(&b, to, privateKey)
	if err != nil {
		return nil, err
	}

	return &protonmail.CalendarImport{
		Cards: []protonmail.Card{*encrypted},
	}, nil
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
	// TODO: handle req

	cal := ical.NewCalendar()
	cal.Version = "2.0"
	cal.ProductID = "-//ProtonMail AG//ProtonMail Calendar//EN"

	// TODO: convert ProtonMail event to iCalendar format
	// This is a simplified implementation
	// In production, you need to properly map all fields

	return &caldav.CalendarObject{
		Path:    formatCalendarPath(event.ID),
		ModTime: event.ModifyTime.Time(),
		ETag:    fmt.Sprintf("%x%x", event.ModifyTime, event.Size),
		Data:    cal,
	}, nil
}

type backend struct {
	c           *protonmail.Client
	cache       map[string]*protonmail.CalendarEvent
	locker      sync.Mutex
	total       int
	privateKeys openpgp.EntityList
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
	return []caldav.Calendar{*calendar}, nil
}

func (b *backend) GetCalendar(ctx context.Context, path string) (*caldav.Calendar, error) {
	if path != calendar.Path {
		return nil, webdav.NewHTTPError(http.StatusNotFound, errors.New("calendar not found"))
	}
	return calendar, nil
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

		// TODO: implement GetCalendarEvent API call
		// event, err = b.c.GetCalendarEvent(id)
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

	// TODO: implement ListCalendarEvents API call
	// For now, return empty list
	return []caldav.CalendarObject{}, nil
}

func (b *backend) QueryCalendarObjects(ctx context.Context, path string, query *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	req := caldav.CalendarDataRequest{AllProp: true}
	if query != nil {
		req = query.DataRequest
	}

	// TODO: optimize
	all, err := b.ListCalendarObjects(ctx, calendar.Path, &req)
	if err != nil {
		return nil, err
	}

	// TODO: implement proper filtering
	return all, nil
}

func (b *backend) PutCalendarObject(ctx context.Context, path string, cal *ical.Calendar, opts *caldav.PutCalendarObjectOptions) (co *caldav.CalendarObject, err error) {
	id, err := parseCalendarPath(path)
	if err != nil {
		return nil, err
	}

	calendarImport, err := formatCalendar(cal, b.privateKeys[0])
	if err != nil {
		return nil, err
	}

	// TODO: implement CreateCalendarEvent and UpdateCalendarEvent API calls
	// For now, return not implemented
	return nil, webdav.NewHTTPError(http.StatusNotImplemented, errors.New("calendar event creation not yet implemented"))
}

func (b *backend) DeleteCalendarObject(ctx context.Context, path string) error {
	id, err := parseCalendarPath(path)
	if err != nil {
		return err
	}

	// TODO: implement DeleteCalendarEvent API call
	return webdav.NewHTTPError(http.StatusNotImplemented, errors.New("calendar event deletion not yet implemented"))
}

func (b *backend) receiveEvents(events <-chan *protonmail.Event) {
	for event := range events {
		b.locker.Lock()
		if event.Refresh&protonmail.EventRefreshCalendar != 0 {
			b.cache = make(map[string]*protonmail.CalendarEvent)
			b.total = -1
		} else if len(event.CalendarEvents) > 0 {
			for _, eventEvent := range event.CalendarEvents {
				switch eventEvent.Action {
				case protonmail.EventCreate:
					if b.total >= 0 {
						b.total++
					}
					fallthrough
				case protonmail.EventUpdate:
					b.cache[eventEvent.ID] = eventEvent.CalendarEvent
				case protonmail.EventDelete:
					delete(b.cache, eventEvent.ID)
					if b.total >= 0 {
						b.total--
					}
				}
			}
		}
		b.locker.Unlock()
	}
}

func NewHandler(c *protonmail.Client, privateKeys openpgp.EntityList, events <-chan *protonmail.Event) http.Handler {
	if len(privateKeys) == 0 {
		panic("hydroxide/caldav: no private key available")
	}

	b := &backend{
		c:           c,
		cache:       make(map[string]*protonmail.CalendarEvent),
		total:       -1,
		privateKeys: privateKeys,
	}

	if events != nil {
		go b.receiveEvents(events)
	}

	return &caldav.Handler{Backend: b}
}
