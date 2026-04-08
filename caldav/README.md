# CalDAV Support for hydroxide

📅 **Status: Implementation in Progress**

This package provides CalDAV calendar sync support for hydroxide, allowing you to sync your ProtonMail calendars with any CalDAV-compatible client.

## Current Status

⚠️ **This is a skeleton implementation.** The core structure is in place, but the following features are still under development:

- [x] Basic CalDAV backend structure
- [x] HTTP handler skeleton
- [x] Event channel integration
- [ ] Calendar event encryption/decryption
- [ ] iCalendar format conversion
- [ ] Full CalDAV protocol support (RFC 4791)
- [ ] Complete CRUD operations

## Usage (When Complete)

```bash
# Run hydroxide as a CalDAV server
hydroxide caldav -addr :8081

# Or run all services together
hydroxide serve
```

Then connect your CalDAV client:
- **Server URL:** `http://localhost:8081`
- **Username:** Your ProtonMail email
- **Password:** Your hydroxide bridge password

## Supported Clients

- Mozilla Thunderbird (with Lightning/Toronto)
- Apple Calendar (macOS, iOS)
- Android devices (via DAVx5)
- Microsoft Outlook (with CalDAV plugin)
- Evolution, Kontact, and other desktop clients

## Implementation Plan

### Phase 1: Core Structure ✅
- Basic backend interface
- HTTP handler skeleton
- Event channel integration

### Phase 2: Encryption/Decryption (TODO)
- ProtonMail calendar event decryption
- iCalendar format encryption
- Key management

### Phase 3: Protocol Support (TODO)
- Full CalDAV protocol (RFC 4791)
- iCalendar format (RFC 5545)
- Calendar queries and filtering

### Phase 4: Testing & Documentation (TODO)
- Unit tests
- Integration tests
- Usage documentation

## Development

### Building

```bash
go build ./cmd/hydroxide
```

### Running Tests

```bash
go test ./caldav/...
```

## Related

- [CardDAV implementation](../carddav/) - Reference for contact sync
- [ProtonMail Calendar API](../protonmail/calendar.go) - Calendar API client
- [CalDAV RFC 4791](https://datatracker.ietf.org/doc/html/rfc4791)
- [iCalendar RFC 5545](https://datatracker.ietf.org/doc/html/rfc5545)

## Bounty

This implementation is part of the CalDAV support bounty:
https://www.bountyhub.dev/en/bounty/view/81479fc7-ca8b-40aa-a117-8a01277e12b0

## Contributing

Contributions are welcome! Areas that need help:

1. **Encryption/Decryption** - Implement ProtonMail calendar event encryption
2. **iCalendar Conversion** - Convert between ProtonMail and iCalendar formats
3. **Protocol Implementation** - Complete CalDAV protocol support
4. **Testing** - Add unit and integration tests

Please open an issue to discuss before submitting PRs.

## License

Same as hydroxide - MIT License

---

**Last Updated:** 2026-04-08  
**Version:** 0.1.0 (skeleton)  
**Status:** 🟡 Implementation in Progress
