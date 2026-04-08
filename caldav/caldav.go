// Package caldav provides CalDAV calendar sync support for hydroxide.
//
// This package implements a CalDAV server that allows syncing ProtonMail
// calendars with any CalDAV-compatible client (Thunderbird, Apple Calendar,
// Android DAVx5, etc.).
//
// Status: Implementation in progress. Core skeleton is in place.
// See README.md for usage information.
package caldav

import (
	"net/http"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/emersion/hydroxide/protonmail"
)

// NewHandler creates a new CalDAV HTTP handler.
//
// Parameters:
//   - c: ProtonMail client
//   - privateKeys: User's decrypted private keys
//   - events: Event channel for real-time updates
//
// Returns an HTTP handler that implements the CalDAV protocol.
//
// TODO: Complete implementation with full CalDAV support:
// - Calendar event encryption/decryption
// - iCalendar format conversion
// - Full CalDAV protocol support (RFC 4791)
// - CRUD operations for calendar events
func NewHandler(c *protonmail.Client, privateKeys openpgp.EntityList, events <-chan *protonmail.Event) http.Handler {
	if len(privateKeys) == 0 {
		panic("hydroxide/caldav: no private key available")
	}

	// TODO: Start event listener for real-time updates
	// TODO: Return proper CalDAV handler
	// For now, return a placeholder handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement full CalDAV protocol handler
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("CalDAV server: Implementation in progress"))
	})
}

// TODO: Implement full CalDAV backend
// - CurrentUserPrincipal
// - CalendarHomeSetPath
// - ListCalendars
// - GetCalendar
// - GetCalendarObject
// - ListCalendarObjects
// - PutCalendarObject
// - DeleteCalendarObject

// TODO: Implement event listener
// func (b *backend) receiveEvents(events <-chan *protonmail.Event) { ... }

// TODO: Implement encryption/decryption helpers
// - DecryptCalendarEvent
// - EncryptCalendarCard
// - Convert to/from iCalendar format

// TODO: Implement full CalDAV protocol support
// - PROPFIND, REPORT, MKCALENDAR, etc.
// - Calendar queries with filtering
// - Free-busy queries
