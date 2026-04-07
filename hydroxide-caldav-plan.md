# CalDAV Support Implementation Plan

## Overview

Implement CalDAV calendar sync support for hydroxide, following the existing CardDAV pattern.

## Reference

Based on existing CardDAV implementation in `carddav/carddav.go`.

## Architecture

```
caldav/
  └── caldav.go      # CalDAV server implementation
cmd/
  └── hydroxide/
      └── main.go    # Add caldav command
```

## Implementation Steps

1. Create `caldav/caldav.go` - CalDAV server
2. Implement calendar sync with ProtonMail API
3. Add caldav command to CLI
4. Write tests
5. Documentation

## Bounty

https://www.bountyhub.dev/en/bounty/view/81479fc7-ca8b-40aa-a117-8a01277e12b0
