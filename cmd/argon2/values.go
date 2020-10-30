package argon2

import (
	"fmt"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/hash"
	"github.com/ory/x/cmdx"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"time"
)

type argon2Config struct {
	c configuration.HasherArgon2Config
}

func (c *argon2Config) HasherArgon2() *configuration.HasherArgon2Config {
	return &c.c
}

func (c *argon2Config) getMemGiB() string {
	return fmt.Sprintf("%.2f GiB", float32(c.c.Memory) / 1024 / 1024)
}

const (
	FlagStartMemory = "start-memory"
	FlagIterations  = "iterations"
	FlagParallelism = "parallelism"
	FlagSaltLength  = "salt-length"
	FlagKeyLength   = "key-length"

	FlagVerbose = "verbose"
	FlagRuns = "probe-runs"
)

func NewValuesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "values [<desired-duration>]",
		Args:  cobra.ExactArgs(1),
		Short: "Calibrate the values for Argon2 so every hashing operation takes the desired time.",
		Long: `Calibrate the configuration values for Argon2 by probing the execution time.
Note that the values depend on the machine you run the hashing on.
When choosing the desired time, UX is in conflict with security. Security should really win out here, therefore we recommend 1s.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			password := []byte("password")
			desiredDuration, err := time.ParseDuration(args[0])
			if err != nil {
				return err
			}
			verbose, err := cmd.Flags().GetBool(FlagVerbose)
			if err != nil {
				return err
			}
			runs, err := cmd.Flags().GetInt(FlagRuns)
			if err != nil {
				return err
			}

			config, err := configFromFlags(cmd)
			if err != nil {
				return err
			}
			hasher := hash.NewHasherArgon2(config)

			var currentDuration time.Duration

			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "Increasing memory to get over %s:\n", desiredDuration)
			}
			for {
				currentDuration, err = probe(cmd, hasher, password, runs, verbose)
				if verbose {
					fmt.Fprintf(cmd.OutOrStdout(), "  took %s with %s of memory\n", currentDuration, config.getMemGiB())
				}

				if currentDuration > desiredDuration {
					break
				}
				// adjust config
				config.c.Memory += 1024*1024
			}

			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "Decreasing memory to get under %s:\n", desiredDuration)
			}
			for {
				currentDuration, err = probe(cmd, hasher, password, runs, verbose)
				if verbose {
					fmt.Fprintf(cmd.OutOrStdout(), "  took %s with %s of memory\n", currentDuration, config.getMemGiB())
				}

				if currentDuration < desiredDuration {
					break
				}
				// adjust config
				config.c.Memory -= 1024*1024
			}
			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "Settled on %s of memory.\n", config.getMemGiB())
				fmt.Fprintf(cmd.OutOrStdout(), "Increasing iterations to get over %s:\n", desiredDuration)
			}
			for {
				currentDuration, err = probe(cmd, hasher, password, runs, verbose)
				if verbose {
					fmt.Fprintf(cmd.OutOrStdout(), "  took %s with %d iterations\n", currentDuration, config.c.Iterations)
				}

				if currentDuration > desiredDuration {
					break
				}
				// adjust config
				config.c.Iterations += 1
			}

			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "Decreasing iterations to get under %s:\n", desiredDuration)
			}
			for {
				currentDuration, err = probe(cmd, hasher, password, runs, verbose)
				if verbose {
					fmt.Fprintf(cmd.OutOrStdout(), "  took %s with %d iterations\n", currentDuration, config.c.Iterations)
				}

				// break also when iterations is 1; this catches the case where 1 was only slightly under the desired time and took longer a bit longer on another run
				if currentDuration < desiredDuration || config.c.Iterations == 1 {
					break
				}
				// adjust config
				config.c.Iterations -= 1
			}
			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "Settled on %d iterations.\n\n", config.c.Iterations)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Memory: %d\nIterations: %d\nLast duration: %s\n", config.c.Memory, config.c.Iterations, currentDuration)

			return nil
		},
	}

	registerArgon2Flags(cmd.Flags())
	cmd.Flags().BoolP(FlagVerbose, "v", false, "Verbose output.")
	cmd.Flags().IntP(FlagRuns, "r", 2, "Runs done for every probing. The median is taken as the result.")

	return cmd
}

func probe(cmd *cobra.Command, hasher hash.Hasher, password []byte, runs int, verbose bool) (time.Duration, error) {
	start := time.Now()

	var mid time.Time
	for i := 0; i < runs; i++ {
		mid = time.Now()
		_, err := hasher.Generate(password)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Could not generate a hash: %s\n", err)
			return 0, cmdx.FailSilently(cmd)
		}
		if verbose {
			fmt.Fprintf(cmd.OutOrStdout(), "    took %s in try %d\n", time.Now().Sub(mid), i)
		}
	}

	return time.Duration(int64(time.Now().Sub(start)) / int64(runs)), nil
}

func registerArgon2Flags(flags *pflag.FlagSet) {
	flags.Uint32P(FlagStartMemory, "m", configuration.Argon2DefaultMemory, "The amount of memory to start probing at in KiB.")
	flags.Uint8(FlagParallelism, configuration.Argon2DefaultParallelism, "The number of threads to use.")
	flags.Uint32P(FlagIterations, "i", 1, "The number of iterations to start probing at.")
	flags.Uint32(FlagSaltLength, configuration.Argon2DefaultSaltLength, "The length of the salt in bytes.")
	flags.Uint32(FlagKeyLength, configuration.Argon2DefaultKeyLength, "The length of the key in bytes.")
}

func configFromFlags(cmd *cobra.Command) (_ *argon2Config, err error) {
	getUint32 := func(name string) (v uint32) {
		// only try getting the flag if there was no error yet
		if err == nil {
			v, err = cmd.Flags().GetUint32(name)
		}
		return v
	}
	getUint8 := func(name string) (v uint8) {
		// only try getting the flag if there was no error yet
		if err == nil {
			v, err = cmd.Flags().GetUint8(name)
		}
		return v
	}

	return &argon2Config{
		c: configuration.HasherArgon2Config{
			Parallelism: getUint8(FlagParallelism),
			Memory:      getUint32(FlagStartMemory),
			Iterations:  getUint32(FlagIterations),
			SaltLength:  getUint32(FlagSaltLength),
			KeyLength:   getUint32(FlagKeyLength),
		},
	}, err
}
