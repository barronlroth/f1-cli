# f1-cli — Product Requirements Document

## Overview

`f1-cli` is a Go CLI that wraps the [OpenF1 API](https://openf1.org) to provide fast, ergonomic access to Formula 1 telemetry, timing, and session data from the terminal. It is designed to be used by both humans and AI agents (coding agents, CXAS tools, OpenClaw skills).

**Why:** No maintained CLI exists for the OpenF1 API. The existing `f1-cli` (Python) used the deprecated Ergast API. This fills the gap with a modern, Go-based tool following best practices from [steipete/discrawl](https://github.com/steipete/discrawl).

## Architecture

- **Language:** Go 1.22+
- **CLI framework:** Cobra
- **Output:** Human-readable table by default, `--json` for machine-readable JSON, `--csv` for CSV
- **No local state:** Pure API wrapper. No database, no config file required.
- **Single binary:** `go build -o f1 ./cmd/f1`
- **Module path:** `github.com/barronlroth/f1-cli`

## API Reference

Base URL: `https://api.openf1.org/v1`

The API uses query parameters for filtering. Comparison operators are supported: `>=`, `<=`, `>`, `<`. The special value `latest` can be used for `session_key` and `meeting_key` to get the most recent.

Rate limits: 3 req/s, 30 req/min (free tier).

### Endpoints (18 total)

| Endpoint | CLI Command | Description |
|---|---|---|
| `/car_data` | `f1 telemetry` | Car telemetry at ~3.7 Hz (speed, RPM, gear, throttle, brake, DRS) |
| `/championship_drivers` | `f1 standings drivers` | Driver championship standings (race sessions only) |
| `/championship_teams` | `f1 standings teams` | Constructor championship standings (race sessions only) |
| `/drivers` | `f1 drivers` | Driver info (name, team, number, acronym) |
| `/intervals` | `f1 intervals` | Gap to leader and car ahead (race only, ~4s updates) |
| `/laps` | `f1 laps` | Lap times, sector times, speed traps |
| `/location` | `f1 location` | Car XYZ position on track (~3.7 Hz) |
| `/meetings` | `f1 meetings` | Grand Prix / test weekend info |
| `/overtakes` | `f1 overtakes` | Position exchanges between drivers |
| `/pit` | `f1 pit` | Pit stop timing and duration |
| `/position` | `f1 positions` | Driver positions throughout session |
| `/race_control` | `f1 race-control` | Flags, safety car, incidents |
| `/sessions` | `f1 sessions` | Session info (FP1, FP2, FP3, Quali, Race, Sprint) |
| `/stints` | `f1 stints` | Tire stints (compound, lap start/end) |
| `/team_radio` | `f1 radio` | Team radio recordings (returns URLs) |
| `/weather` | `f1 weather` | Track/air temp, humidity, wind, rain |

Note: `/laps_count`, `/car_data` aggregation, and any future endpoints should be easy to add due to the generic architecture.

## Command Design

### Global Flags

```
--json              Output as JSON
--csv               Output as CSV
--session KEY       Session key (number or "latest")
--meeting KEY       Meeting key (number or "latest")
--driver DRIVER     Driver number or 3-letter acronym (VER, HAM, NOR, etc.)
--limit N           Limit number of results
--help              Show help
--version           Show version
```

### Driver Resolution

The `--driver` flag accepts either:
- A driver number (e.g., `1`, `44`, `4`)
- A 3-letter acronym (e.g., `VER`, `HAM`, `NOR`)

When an acronym is provided, the CLI resolves it to a driver number by querying `/drivers?session_key=latest&name_acronym=VER`. This resolution is cached for the duration of the command.

### Example Commands

```bash
# List all drivers in the current session
f1 drivers --session latest

# Get Verstappen's lap times in the latest session
f1 laps --session latest --driver VER

# Get fastest pit stops (under 2.5 seconds)
f1 pit --session latest --filter "stop_duration<2.5"

# Championship standings after latest race
f1 standings drivers --session latest

# Weather during latest session
f1 weather --session latest --limit 5

# Race control messages (flags, safety car)
f1 race-control --session latest

# Team radio for a specific driver
f1 radio --session latest --driver HAM

# Find a specific meeting
f1 meetings --year 2025 --country Singapore

# Sessions for a meeting
f1 sessions --meeting latest

# Telemetry data as JSON for piping
f1 telemetry --session latest --driver VER --json | jq '.[] | .speed'

# Overtakes in a race
f1 overtakes --session latest

# Tire stints
f1 stints --session latest --driver VER

# Position changes
f1 positions --session latest --driver NOR

# Intervals / gaps
f1 intervals --session latest --limit 20
```

### The `--filter` Flag

For advanced filtering, support a generic `--filter` flag that maps directly to API query params:

```bash
f1 telemetry --session 9159 --driver 55 --filter "speed>=315"
f1 laps --session 9161 --filter "lap_duration<90"
f1 pit --session latest --filter "stop_duration<2.3"
```

Multiple filters can be chained: `--filter "speed>=300" --filter "throttle>=95"`

### Subcommand: `f1 doctor`

Sanity check command:
- Verify API is reachable
- Show current rate limit status (if headers available)
- Show latest session info
- Print version

### Subcommand: `f1 standings`

Has two sub-subcommands:
- `f1 standings drivers` → `/championship_drivers`
- `f1 standings teams` → `/championship_teams`

When called without a sub-subcommand, default to `drivers`.

## Output Formatting

### Table Output (default)

Use a clean, aligned table format. Example for `f1 drivers --session latest`:

```
#   DRIVER  NAME              TEAM
1   VER     Max Verstappen    Red Bull Racing
4   NOR     Lando Norris      McLaren
44  HAM     Lewis Hamilton    Ferrari
...
```

For `f1 laps --session latest --driver VER --limit 3`:

```
LAP  DURATION  S1      S2      S3      SPEED TRAP
1    93.412    27.1    39.2    27.1    298
2    91.823    26.8    38.6    26.4    305
3    91.654    26.7    38.5    26.4    307
```

### JSON Output

Pretty-printed JSON array, exactly as returned from the API. No transformation.

### CSV Output

Standard CSV with headers. Direct from API with `csv=true` param.

## Implementation Plan

### Phase 1: Core Infrastructure
1. Go module init, Cobra root command, version
2. HTTP client with rate limiting (3 req/s) and retry on 429
3. Generic API query builder (endpoint + params + comparison operators)
4. Output formatters (table, JSON, CSV)
5. Driver acronym → number resolver

### Phase 2: All Commands
6. `f1 drivers`
7. `f1 sessions`
8. `f1 meetings`
9. `f1 laps`
10. `f1 telemetry`
11. `f1 pit`
12. `f1 positions`
13. `f1 intervals`
14. `f1 standings drivers` / `f1 standings teams`
15. `f1 weather`
16. `f1 race-control`
17. `f1 radio`
18. `f1 stints`
19. `f1 overtakes`
20. `f1 location`
21. `f1 doctor`

### Phase 3: Polish
22. `--filter` flag for arbitrary API params
23. Comprehensive `--help` for every command with examples
24. README.md with install instructions, usage examples
25. CI: `go build`, `go test`, `golangci-lint`

## Non-Goals

- No local caching/storage
- No real-time streaming/websocket support (API doesn't support it)
- No authentication management (API is free for historical data)
- No TUI/interactive mode
- No data visualization (that's what piping to other tools is for)

## Success Criteria

- `go build` produces a single binary
- All 18 API endpoints have corresponding commands
- `--json`, `--csv`, and table output work for every command
- `--driver VER` resolves acronyms correctly
- `--session latest` works as expected
- `f1 doctor` confirms API connectivity
- README with clear install + usage docs
- Tests for driver resolution, query building, output formatting

## Reference

- OpenF1 API docs: https://openf1.org/docs
- OpenF1 GitHub: https://github.com/br-g/openf1
- Architecture reference: https://github.com/steipete/discrawl
- Rate limits: 3 req/s, 30 req/min (free tier)
