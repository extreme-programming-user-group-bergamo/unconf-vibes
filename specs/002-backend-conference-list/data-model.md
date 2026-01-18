# Phase 1: Data Model

**Feature**: 002-backend-conference-list

## Entities

### Conference

*The primary entity representing an event.*

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Unique identifier (e.g., "socrates-2026") |
| `Name` | `string` | Display name of the conference |
| `StartDate` | `date` | ISO-8601 start date |
| `EndDate` | `date` | ISO-8601 end date |
| `Location` | `string` | City/Venue description |
| `Status` | `string` | Current state: `open` or `closed` |

## Storage Schema

**Mock Source**: `internal/backend/data/conferences.json`

```json
[
  {
    "id": "socrates-26",
    "name": "SoCraTes IT 2026",
    "startDate": "2026-05-20",
    "endDate": "2026-05-22",
    "location": "Rimini, Italy",
    "status": "open"
  }
]
```
