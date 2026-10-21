package app

import (
	fd "github.com/flashdatadev/flashdata-go"
	"github.com/spf13/cobra"
	"time"
)

func (a *App) addAccount() {
	account := &cobra.Command{Use: "account", Short: "Read shared balance, source pricing, limits and usage"}
	a.Root.AddCommand(account)
	for _, action := range []string{"balance", "pricing", "limits", "usage"} {
		action := action
		var source, from, to string
		var count int
		cmd := &cobra.Command{Use: action, Short: "Read account " + action + " without submitting paid work", Args: cobra.NoArgs}
		if action == "pricing" || action == "usage" {
			cmd.Flags().StringVar(&source, "source", "", "filter by source")
		}
		if action == "pricing" {
			cmd.Flags().IntVar(&count, "request-count", 1, "estimate this many requests")
		}
		if action == "usage" {
			cmd.Flags().StringVar(&from, "from", "", "RFC3339 start time")
			cmd.Flags().StringVar(&to, "to", "", "RFC3339 end time")
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			ctx, cancel := a.context(cmd)
			defer cancel()
			s, err := a.store()
			if err != nil {
				return err
			}
			v, err := a.load(ctx, s)
			if err != nil {
				return err
			}
			auth, err := a.authorize(ctx, v, fd.Source(source), false)
			if err != nil {
				return err
			}
			var result any
			switch action {
			case "balance":
				result, err = auth.Client.Account.Balance(ctx)
			case "limits":
				result, err = auth.Client.Account.Limits(ctx)
			case "pricing":
				result, err = auth.Client.Account.Pricing(ctx, fd.PricingParams{Source: fd.Source(source), RequestCount: count})
			case "usage":
				p := fd.UsageParams{Source: fd.Source(source)}
				if from != "" {
					p.From, err = time.Parse(time.RFC3339, from)
					if err != nil {
						return fail("invalid_time", "from must be RFC3339.", "")
					}
				}
				if to != "" {
					p.To, err = time.Parse(time.RFC3339, to)
					if err != nil {
						return fail("invalid_time", "to must be RFC3339.", "")
					}
				}
				result, err = auth.Client.Account.Usage(ctx, p)
			}
			if err != nil {
				return err
			}
			return a.success(result, "", false, nil)
		}
		account.AddCommand(cmd)
	}
}
