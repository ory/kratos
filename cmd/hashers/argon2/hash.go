package argon2

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/hash"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/configx"
	"github.com/ory/x/flagx"
)

const (
	FlagBench    = "bench"
	FlagParallel = "parallel"
)

func newHashCmd() *cobra.Command {
	flagConfig := &argon2Config{}

	cmd := &cobra.Command{
		Use:  "hash <password1> [<password2> ...]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := configProvider(cmd, flagConfig)
			if err != nil {
				return err
			}

			hasher := hash.NewHasherArgon2(conf)
			hashes := make([][]byte, len(args))
			errs := make(chan error, len(args))

			start := time.Now()

			for i, pw := range args {
				go func(i int, pw string) {
					start := time.Now()
					h, err := hasher.Generate(cmd.Context(), []byte(pw))
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", time.Since(start))

					hashes[i] = h
					errs <- err
				}(i, pw)

				if !flagx.MustGetBool(cmd, FlagParallel) {
					if err := <-errs; err != nil {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Could not generate hash: %s\n", err.Error())
						return cmdx.FailSilently(cmd)
					}
				}
			}

			if flagx.MustGetBool(cmd, FlagParallel) {
				for i := 0; i < len(args); i++ {
					if err := <-errs; err != nil {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Could not generate hash: %s\n", err.Error())
						return cmdx.FailSilently(cmd)
					}
				}
			}

			if flagx.MustGetBool(cmd, FlagBench) {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", time.Since(start))
				return nil
			}

			for _, h := range hashes {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", h)
			}
			return nil
		},
	}

	flags := cmd.Flags()

	flags.Bool(FlagBench, false, "Run the hashing as a benchmark.")
	flags.Bool(FlagParallel, false, "Run all hashing operations in parallel.")

	registerArgon2ConfigFlags(flags, flagConfig)
	configx.RegisterFlags(flags)

	return cmd
}
