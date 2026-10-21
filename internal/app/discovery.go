package app

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/flashdatadev/flashdata-cli/internal/state"
	fd "github.com/flashdatadev/flashdata-go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (a *App) discoveryClient(ctx context.Context) (*fd.Client, error) {
	s, err := a.store()
	if err != nil {
		return nil, err
	}
	v, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	return a.client(v, "")
}
func (a *App) addDiscovery() {
	discover := &cobra.Command{Use: "discover", Short: "Read service health, onboarding capabilities and documentation", Args: cobra.NoArgs}
	discover.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := a.context(cmd)
		defer cancel()
		c, err := a.discoveryClient(ctx)
		if err != nil {
			return err
		}
		health, err := c.Public.Health(ctx)
		if err != nil {
			return err
		}
		capabilities, capErr := c.Onboarding.Capabilities(ctx)
		var api *fd.APIError
		if capErr != nil && (!errors.As(capErr, &api) || api.StatusCode != 404) {
			return capErr
		}
		return a.success(map[string]any{"service": health, "onboarding": capabilities, "onboardingAvailable": capErr == nil && capabilities.Available, "docs": "https://flashdata.dev/docs", "account": "https://flashdata.dev", "queryProtocol": "api_key", "oauthBoundary": "MCP OAuth tokens are not REST API keys."}, "", false, nil)
	}
	a.Root.AddCommand(discover)
	sources := &cobra.Command{Use: "sources", Short: "Discover available, permitted and CLI-supported sources"}
	a.Root.AddCommand(sources)
	list := &cobra.Command{Use: "list", Short: "List source support and selected-key permissions", Args: cobra.NoArgs}
	list.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := a.context(cmd)
		defer cancel()
		c, err := a.discoveryClient(ctx)
		if err != nil {
			return err
		}
		catalog, err := c.Public.Sources(ctx)
		if err != nil {
			return err
		}
		s, err := a.store()
		if err != nil {
			return err
		}
		v, err := a.load(ctx, s)
		var auth *access
		if err == nil {
			auth, err = a.authorize(ctx, v, "", false)
			if err != nil {
				return err
			}
		} else if !errors.Is(err, state.ErrMissing) {
			return err
		}
		data := []any{}
		for _, entry := range catalog {
			_, schemaErr := fd.QuerySchema(fd.Source(entry.ID))
			permitted := false
			verified := auth != nil
			if auth != nil {
				for _, source := range auth.Sources {
					permitted = permitted || string(source) == entry.ID
				}
			}
			data = append(data, map[string]any{"source": entry.ID, "name": entry.Name, "available": true, "supported": schemaErr == nil, "permitted": permitted, "permissionVerified": verified, "modes": entry.Modes})
		}
		return a.success(data, "", false, nil)
	}
	sources.AddCommand(list)
	show := &cobra.Command{Use: "show SOURCE", Short: "Show schema, defaults, modes, example and optional account pricing", Args: cobra.ExactArgs(1)}
	show.RunE = func(cmd *cobra.Command, args []string) error {
		source := fd.Source(args[0])
		schema, err := fd.QuerySchema(source)
		if err != nil {
			return err
		}
		// Reflect the SDK's runtime additions to its embedded source schemas:
		// strict top-level fields and optional background-only callbacks.
		var querySchema map[string]json.RawMessage
		_ = json.Unmarshal(schema, &querySchema)
		var properties map[string]json.RawMessage
		_ = json.Unmarshal(querySchema["properties"], &properties)
		properties["callback_url"] = json.RawMessage(`{"type":"string","maxLength":2048,"description":"Optional public HTTPS callback for async/batch only; nonempty URLs use port 443 and cannot contain credentials or fragments."}`)
		querySchema["properties"], _ = json.Marshal(properties)
		querySchema["additionalProperties"] = json.RawMessage(`false`)
		schema, _ = json.Marshal(querySchema)
		ctx, cancel := a.context(cmd)
		defer cancel()
		modes := []string{"realtime", "async", "batch"}
		if source == fd.SourceYouTubeDownload {
			modes = []string{"async", "batch"}
		}
		example := map[string]any{"source": source}
		if source == fd.SourceYouTubeChannel {
			example["handle"] = "@GoogleDevelopers"
		} else if source == fd.SourceGoogleSearch || source == fd.SourceGoogleImages || source == fd.SourceGoogleNews || source == fd.SourceGoogleShopping || source == fd.SourceGooglePlaces || source == fd.SourceGoogleVideos || source == fd.SourceYouTubeSearch || source == fd.SourceYouTubeSearchExtended || source == fd.SourceYouTubeSuggestions {
			example["query"] = "machine learning"
		} else {
			example["video_id"] = "dQw4w9WgXcQ"
		}
		if source == fd.SourceYouTubeDownload {
			example["params"] = map[string]any{"storage": map[string]any{"storage_type": "s3", "storage_url": "YOUR_BUCKET", "access_key": "YOUR_ACCESS_KEY", "secret_key": "YOUR_SECRET_KEY"}, "download_type": "audio", "output_format": "mp3"}
		}
		var price any
		var permitted *bool
		s, err := a.store()
		if err != nil {
			return err
		}
		v, e := a.load(ctx, s)
		if e == nil {
			auth, e := a.authorize(ctx, v, "", false)
			if e != nil {
				return e
			}
			allowed := requireSource(auth, source) == nil
			permitted = &allowed
			if allowed {
				p, e := auth.Client.Account.Pricing(ctx, fd.PricingParams{Source: source})
				if e == nil {
					price = p
				}
			}
		} else if !errors.Is(e, state.ErrMissing) {
			return e
		}
		data := a.redacted(map[string]any{"source": source, "modes": modes, "example": example, "requiresCustomerStorage": source == fd.SourceYouTubeDownload, "pricing": price, "permitted": permitted, "permissionVerified": permitted != nil, "docs": "https://flashdata.dev/docs/sources/" + string(source), "flatParametersAccepted": true}).(map[string]any)
		data["inputSchema"] = json.RawMessage(schema)
		return a.writeSuccess(data, "", false, nil)
	}
	sources.AddCommand(show)
	commands := &cobra.Command{Use: "commands", Short: "Return the command tree for agent discovery", Args: cobra.NoArgs}
	commands.RunE = func(cmd *cobra.Command, args []string) error { return a.success(commandTree(a.Root), "", false, nil) }
	a.Root.AddCommand(commands)
	var offline bool
	doctor := &cobra.Command{Use: "doctor", Short: "Diagnose storage and access without paid queries or account mutations", Args: cobra.NoArgs}
	doctor.Flags().BoolVar(&offline, "offline", false, "inspect local storage only")
	doctor.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := a.context(cmd)
		defer cancel()
		s, err := a.store()
		if err != nil {
			return err
		}
		diagnostics := map[string]any{"profile": s.Profile, "stateDirectory": s.Root, "offline": offline, "paidQueriesSubmitted": 0, "registrationAttempted": false}
		v, err := a.load(ctx, s)
		if err != nil {
			diagnostics["credentialsAvailable"] = false
			diagnostics["problem"] = a.classify(err)
			return &Error{Code: "doctor_failed", Message: "Credentials or private storage need attention.", Next: "Run auth login or explicitly recover the original store; no trial was created.", Details: diagnostics}
		}
		protection, _ := s.Protection()
		diagnostics["credentialsAvailable"] = true
		diagnostics["credentialProtection"] = protection
		if offline {
			return a.success(diagnostics, "", false, nil)
		}
		auth, err := a.authorize(ctx, v, "", false)
		if err != nil {
			diagnostics["problem"] = a.classify(err)
			return &Error{Code: "doctor_failed", Message: "Access verification failed.", Details: diagnostics}
		}
		diagnostics["identity"] = publicIdentity(auth)
		diagnostics["verifiedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
		diagnostics["healthy"] = true
		diagnostics["recurringReady"] = auth.Identity.RecurringAllowed
		return a.success(diagnostics, "", false, nil)
	}
	a.Root.AddCommand(doctor)
}
func commandTree(cmd *cobra.Command) any {
	flags := []any{}
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			flags = append(flags, map[string]any{"name": f.Name, "type": f.Value.Type(), "default": f.DefValue, "description": f.Usage})
		}
	})
	children := []any{}
	commands := cmd.Commands()
	sort.Slice(commands, func(i, j int) bool { return commands[i].Name() < commands[j].Name() })
	for _, child := range commands {
		if child.Hidden {
			continue
		}
		children = append(children, commandTree(child))
	}
	return map[string]any{"name": cmd.Name(), "use": cmd.Use, "description": cmd.Short, "flags": flags, "commands": children}
}
