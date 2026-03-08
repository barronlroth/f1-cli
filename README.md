# f1-cli 🏎️

A fast, ergonomic CLI for the [OpenF1 API](https://openf1.org) — Formula 1 telemetry, timing, and session data from your terminal.

## Install

### Homebrew (recommended)

```bash
brew tap barronlroth/tap
brew install barronlroth/tap/f1-cli
```

### From source

Requires Go 1.22+:

```bash
git clone https://github.com/barronlroth/f1-cli.git
cd f1-cli
go build -o f1 ./cmd/f1
```

## Usage

```bash
f1 --help
```

### Global flags

```
--json              Output as JSON
--csv               Output as CSV
--session KEY       Session key (number or "latest")
--meeting KEY       Meeting key (number or "latest")
--driver DRIVER     Driver number or 3-letter acronym (VER, HAM, NOR)
--limit N           Limit number of results (applied client-side)
--filter EXPR       Raw API filter (repeatable), e.g. speed>=300
```

## Commands

| Command | API Endpoint | Description |
|---|---|---|
| `f1 drivers` | `/drivers` | Driver info (name, team, number) |
| `f1 sessions` | `/sessions` | Session list (FP1-3, Quali, Sprint, Race) |
| `f1 meetings` | `/meetings` | Grand Prix weekends |
| `f1 laps` | `/laps` | Lap times, sector times, speed traps |
| `f1 telemetry` | `/car_data` | Speed, RPM, gear, throttle, brake, DRS |
| `f1 pit` | `/pit` | Pit stop timing and duration |
| `f1 positions` | `/position` | Position changes throughout session |
| `f1 intervals` | `/intervals` | Gap to leader and car ahead |
| `f1 standings drivers` | `/championship_drivers` | Driver championship standings |
| `f1 standings teams` | `/championship_teams` | Constructor standings |
| `f1 weather` | `/weather` | Track/air temp, humidity, wind, rain |
| `f1 race-control` | `/race_control` | Flags, safety car, incidents |
| `f1 radio` | `/team_radio` | Team radio recording URLs |
| `f1 stints` | `/stints` | Tire compound and stint laps |
| `f1 overtakes` | `/overtakes` | Position exchanges between drivers |
| `f1 location` | `/location` | Car XYZ position on track |
| `f1 doctor` | — | API connectivity check |

## Examples

```bash
# Drivers in current session
f1 drivers --session latest

# Verstappen lap data by acronym
f1 laps --session latest --driver VER

# Fast pit stops
f1 pit --session latest --filter "stop_duration<2.5"

# Championship standings
f1 standings drivers --session latest

# Weather during session
f1 weather --session latest --limit 5

# Race control messages
f1 race-control --session latest

# Team radio for Hamilton
f1 radio --session latest --driver HAM

# Find a Grand Prix
f1 meetings --year 2025 --country Singapore

# Telemetry as JSON for piping
f1 telemetry --session latest --driver VER --json | jq '.[0].speed'

# Overtakes
f1 overtakes --session latest

# Tire strategy
f1 stints --session latest --driver VER

# Health check
f1 doctor
```

## API

- **Data:** 2023 season onwards, no auth required
- **Rate limits:** 3 req/s, 30 req/min (free tier) — handled automatically
- **Formats:** Table (default), JSON (`--json`), CSV (`--csv`)
- **Driver resolution:** `--driver VER` auto-resolves acronyms to driver numbers

## License

MIT
