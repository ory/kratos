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
	FlagParallel = "parallel"
)

func newHashCmd() *cobra.Command {
	flagConfig := &argon2Config{}

	cmd := &cobra.Command{
		Use:   "hash <password1> [<password2> ...]",
		Short: "Hash a list of passwords for benchmarking the hashing parameters.",
		Args:  cobra.MinimumNArgs(1),
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
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "password %d: %s\n", i, time.Since(start))

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

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "total: %s\n", time.Since(start))
			return nil
		},
	}

	flags := cmd.Flags()

	flags.Bool(FlagParallel, false, "Run all hashing operations in parallel.")

	registerArgon2ConfigFlags(flags, flagConfig)
	configx.RegisterFlags(flags)

	return cmd
}
