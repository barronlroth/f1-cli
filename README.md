# f1-cli

`f1-cli` is a Go command-line client for the [OpenF1 API](https://openf1.org).

## Requirements

- Go 1.22+

## Install / Build

```bash
git clone https://github.com/barronlroth/f1-cli.git
cd f1-cli
go build -o f1 ./cmd/f1
```

## Usage

```bash
./f1 --help
```

Global flags:

```text
--json              Output as JSON
--csv               Output as CSV
--session KEY       Session key (number or "latest")
--meeting KEY       Meeting key (number or "latest")
--driver DRIVER     Driver number or 3-letter acronym (VER, HAM, NOR)
--limit N           Limit number of results
--filter EXPR       Raw API filter (repeatable), e.g. speed>=300
```

## Commands

- `f1 drivers` -> `/drivers`
- `f1 sessions` -> `/sessions`
- `f1 meetings` -> `/meetings`
- `f1 laps` -> `/laps`
- `f1 telemetry` -> `/car_data`
- `f1 pit` -> `/pit`
- `f1 positions` -> `/position`
- `f1 intervals` -> `/intervals`
- `f1 standings drivers` -> `/championship_drivers`
- `f1 standings teams` -> `/championship_teams`
- `f1 weather` -> `/weather`
- `f1 race-control` -> `/race_control`
- `f1 radio` -> `/team_radio`
- `f1 stints` -> `/stints`
- `f1 overtakes` -> `/overtakes`
- `f1 location` -> `/location`
- `f1 doctor` -> connectivity and API status checks

## Examples

```bash
# Drivers in current session
./f1 drivers --session latest

# Verstappen lap data by acronym resolution
./f1 laps --session latest --driver VER

# Fast pit stops with comparison filter
./f1 pit --session latest --filter "stop_duration<2.5"

# Latest championship standings
./f1 standings drivers --session latest

# Session weather samples
./f1 weather --session latest --limit 5

# Race control messages
./f1 race-control --session latest

# Team radio for Hamilton
./f1 radio --session latest --driver HAM

# Meeting lookup
./f1 meetings --year 2025 --country Singapore

# Sessions in latest meeting
./f1 sessions --meeting latest

# Telemetry as JSON
./f1 telemetry --session latest --driver VER --json

# Overtakes in latest session
./f1 overtakes --session latest

# Tire stints
./f1 stints --session latest --driver VER

# Position changes
./f1 positions --session latest --driver NOR

# Intervals and gaps
./f1 intervals --session latest --limit 20

# Health check
./f1 doctor
```
