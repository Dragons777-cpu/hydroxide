# CalDAV Support

Basic CalDAV calendar sync for hydroxide.

## Status

Work in progress. Currently supports:
- Listing calendars
- Reading events
- iCalendar export

Write support (create/update/delete events) is not yet implemented.

## Usage

After starting hydroxide, the CalDAV server will be available at:
```
http://localhost:8080/caldav/
```

Configure your CalDAV client (Thunderbird, Evolution, etc.) to use this URL.

## Implementation Notes

This implementation follows the existing CardDAV pattern in hydroxide.
The CalDAV backend wraps the ProtonMail calendar API and exposes it
through the standard CalDAV protocol.

## TODO

- [ ] Implement event creation
- [ ] Implement event updates
- [ ] Implement event deletion
- [ ] Add proper error handling
- [ ] Add authentication middleware
- [ ] Support recurring events
- [ ] Support event reminders

## Testing

```bash
go test ./caldav/...
```

## References

- RFC 4791: CalDAV
- RFC 5545: iCalendar
- go-webdav/caldav package
