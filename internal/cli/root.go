package cli

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/barronlroth/f1-cli/internal/client"
	"github.com/barronlroth/f1-cli/internal/driver"
	"github.com/barronlroth/f1-cli/internal/output"
	"github.com/spf13/cobra"
)

type outputFormat string

const (
	formatTable outputFormat = "table"
	formatJSON  outputFormat = "json"
	formatCSV   outputFormat = "csv"
)

type globalOptions struct {
	jsonOut bool
	csvOut  bool
	session string
	meeting string
	driver  string
	limit   int
	filters []string
}

type app struct {
	version  string
	out      io.Writer
	errOut   io.Writer
	opts     *globalOptions
	client   *client.Client
	resolver *driver.Resolver
}

type endpointSpec struct {
	use         string
	aliases     []string
	short       string
	long        string
	example     string
	endpoint    string
	configure   func(cmd *cobra.Command, local *endpointLocalOptions)
	buildParams func(local *endpointLocalOptions) map[string]string
}

type endpointLocalOptions struct {
	year    int
	country string
}

func Execute(version string, out, errOut io.Writer) error {
	cmd, cleanup := NewRootCmd(version, out, errOut)
	defer cleanup()
	return cmd.Execute()
}

func NewRootCmd(version string, out, errOut io.Writer) (*cobra.Command, func()) {
	opts := &globalOptions{}
	apiClient := client.New(client.DefaultBaseURL, nil)
	resolver := driver.NewResolver(apiClient)

	a := &app{
		version:  version,
		out:      out,
		errOut:   errOut,
		opts:     opts,
		client:   apiClient,
		resolver: resolver,
	}

	rootCmd := &cobra.Command{
		Use:           "f1",
		Short:         "Formula 1 data from OpenF1",
		Long:          "f1-cli is a terminal interface for OpenF1 telemetry, timing, and session APIs.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := a.selectedFormat(); err != nil {
				return err
			}
			return validateFilters(a.opts.filters)
		},
	}

	rootCmd.SetOut(out)
	rootCmd.SetErr(errOut)
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	rootCmd.PersistentFlags().BoolVar(&opts.jsonOut, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&opts.csvOut, "csv", false, "Output as CSV")
	rootCmd.PersistentFlags().StringVar(&opts.session, "session", "", `Session key (number or "latest")`)
	rootCmd.PersistentFlags().StringVar(&opts.meeting, "meeting", "", `Meeting key (number or "latest")`)
	rootCmd.PersistentFlags().StringVar(&opts.driver, "driver", "", "Driver number or 3-letter acronym (VER, HAM, NOR)")
	rootCmd.PersistentFlags().IntVar(&opts.limit, "limit", 0, "Limit number of results")
	rootCmd.PersistentFlags().StringArrayVar(&opts.filters, "filter", nil, "Raw API filter, can be repeated (e.g. speed>=300)")

	rootCmd.AddCommand(
		a.newEndpointCommand(endpointSpec{
			use:      "drivers",
			short:    "Driver info (name, team, number, acronym)",
			endpoint: "drivers",
			example:  "f1 drivers --session latest",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "sessions",
			short:    "Session info (FP1, FP2, FP3, Quali, Race, Sprint)",
			endpoint: "sessions",
			example:  "f1 sessions --meeting latest",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "meetings",
			short:    "Grand Prix or test weekend info",
			endpoint: "meetings",
			example:  "f1 meetings --year 2025 --country Singapore",
			configure: func(cmd *cobra.Command, local *endpointLocalOptions) {
				cmd.Flags().IntVar(&local.year, "year", 0, "Filter by year")
				cmd.Flags().StringVar(&local.country, "country", "", "Filter by country name")
			},
			buildParams: func(local *endpointLocalOptions) map[string]string {
				params := map[string]string{}
				if local.year > 0 {
					params["year"] = strconv.Itoa(local.year)
				}
				if strings.TrimSpace(local.country) != "" {
					params["country_name"] = local.country
				}
				return params
			},
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "laps",
			short:    "Lap times, sectors, and speed trap data",
			endpoint: "laps",
			example:  "f1 laps --session latest --driver VER --limit 5",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "telemetry",
			aliases:  []string{"car-data"},
			short:    "Car telemetry (speed, RPM, gear, throttle, brake, DRS)",
			endpoint: "car_data",
			example:  "f1 telemetry --session latest --driver VER --filter \"speed>=315\"",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "pit",
			short:    "Pit stop timing and duration",
			endpoint: "pit",
			example:  "f1 pit --session latest --filter \"stop_duration<2.5\"",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "positions",
			aliases:  []string{"position"},
			short:    "Driver positions throughout session",
			endpoint: "position",
			example:  "f1 positions --session latest --driver NOR",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "intervals",
			short:    "Gap to leader and car ahead",
			endpoint: "intervals",
			example:  "f1 intervals --session latest --limit 20",
		}),
		a.newStandingsCommand(),
		a.newEndpointCommand(endpointSpec{
			use:      "weather",
			short:    "Track and air weather metrics",
			endpoint: "weather",
			example:  "f1 weather --session latest --limit 5",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "race-control",
			aliases:  []string{"race_control"},
			short:    "Race control messages (flags, incidents, safety car)",
			endpoint: "race_control",
			example:  "f1 race-control --session latest",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "radio",
			aliases:  []string{"team-radio"},
			short:    "Team radio recordings (URL metadata)",
			endpoint: "team_radio",
			example:  "f1 radio --session latest --driver HAM",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "stints",
			short:    "Tire stint data by driver",
			endpoint: "stints",
			example:  "f1 stints --session latest --driver VER",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "overtakes",
			short:    "Overtake events during a session",
			endpoint: "overtakes",
			example:  "f1 overtakes --session latest",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "location",
			short:    "Car XYZ track location data",
			endpoint: "location",
			example:  "f1 location --session latest --driver 1 --limit 10",
		}),
		a.newDoctorCommand(),
	)

	cleanup := func() {
		a.client.Close()
	}

	return rootCmd, cleanup
}

func (a *app) newEndpointCommand(spec endpointSpec) *cobra.Command {
	local := &endpointLocalOptions{}

	cmd := &cobra.Command{
		Use:     spec.use,
		Aliases: spec.aliases,
		Short:   spec.short,
		Long:    spec.long,
		Example: spec.example,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := a.commonParams(cmd.Context())
			if err != nil {
				return err
			}

			if spec.buildParams != nil {
				for key, value := range spec.buildParams(local) {
					if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
						continue
					}
					params[key] = value
				}
			}

			return a.executeEndpoint(cmd.Context(), spec.endpoint, params)
		},
	}

	if spec.configure != nil {
		spec.configure(cmd, local)
	}

	return cmd
}

func (a *app) newStandingsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "standings",
		Short: "Championship standings",
		Long:  "Driver and constructor championship standings for race sessions.",
		Example: strings.Join([]string{
			"f1 standings",
			"f1 standings drivers --session latest",
			"f1 standings teams --session latest",
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := a.commonParams(cmd.Context())
			if err != nil {
				return err
			}
			return a.executeEndpoint(cmd.Context(), "championship_drivers", params)
		},
	}

	cmd.AddCommand(
		a.newEndpointCommand(endpointSpec{
			use:      "drivers",
			short:    "Driver championship standings",
			endpoint: "championship_drivers",
			example:  "f1 standings drivers --session latest",
		}),
		a.newEndpointCommand(endpointSpec{
			use:      "teams",
			short:    "Constructor championship standings",
			endpoint: "championship_teams",
			example:  "f1 standings teams --session latest",
		}),
	)

	return cmd
}

func (a *app) selectedFormat() (outputFormat, error) {
	if a.opts.jsonOut && a.opts.csvOut {
		return "", errors.New("only one of --json or --csv may be set")
	}
	if a.opts.csvOut {
		return formatCSV, nil
	}
	if a.opts.jsonOut {
		return formatJSON, nil
	}
	return formatTable, nil
}

func (a *app) commonParams(ctx context.Context) (map[string]string, error) {
	params := map[string]string{}

	if session := strings.TrimSpace(a.opts.session); session != "" {
		params["session_key"] = session
	}
	if meeting := strings.TrimSpace(a.opts.meeting); meeting != "" {
		params["meeting_key"] = meeting
	}
	// Note: limit is applied client-side, not sent to the API.
	// OpenF1 does not support a "limit" query parameter.

	if driverInput := strings.TrimSpace(a.opts.driver); driverInput != "" {
		resolutionSession := strings.TrimSpace(a.opts.session)
		if resolutionSession == "" {
			resolutionSession = "latest"
		}

		driverNumber, err := a.resolver.Resolve(ctx, driverInput, resolutionSession)
		if err != nil {
			return nil, err
		}
		params["driver_number"] = driverNumber
	}

	return params, nil
}

func (a *app) executeEndpoint(ctx context.Context, endpoint string, params map[string]string) error {
	format, err := a.selectedFormat()
	if err != nil {
		return err
	}

	switch format {
	case formatCSV:
		csvParams := copyMap(params)
		csvParams["csv"] = "true"

		resp, err := a.client.GetRaw(ctx, endpoint, csvParams, a.opts.filters)
		if err != nil {
			return err
		}
		return output.WriteCSV(a.out, resp.Body)
	case formatJSON:
		records, err := a.client.Query(ctx, endpoint, params, a.opts.filters)
		if err != nil {
			return err
		}
		if a.opts.limit > 0 && len(records) > a.opts.limit {
			records = records[:a.opts.limit]
		}
		return output.WriteJSONRecords(a.out, records)
	default:
		records, err := a.client.Query(ctx, endpoint, params, a.opts.filters)
		if err != nil {
			return err
		}
		if a.opts.limit > 0 && len(records) > a.opts.limit {
			records = records[:a.opts.limit]
		}
		return output.WriteTable(a.out, records)
	}
}

func copyMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func validateFilters(filters []string) error {
	for _, filter := range filters {
		if strings.TrimSpace(filter) == "" {
			return errors.New("filter cannot be empty")
		}
	}
	return nil
}
