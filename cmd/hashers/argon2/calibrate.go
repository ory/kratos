package argon2

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/hash"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
)

type (
	argon2Config struct {
		c configuration.HasherArgon2Config
	}
)

func (c *argon2Config) HasherArgon2() *configuration.HasherArgon2Config {
	return &c.c
}

func (c *argon2Config) getMemFormat() string {
	return (datasize.ByteSize(c.c.Memory) * datasize.KB).HumanReadable()
}

const (
	FlagStartMemory     = "start-memory"
	FlagMaxMemory       = "max-memory"
	FlagAdjustMemory    = "adjust-memory-by"
	FlagStartIterations = "start-iterations"
	FlagParallelism     = "parallelism"
	FlagSaltLength      = "salt-length"
	FlagKeyLength       = "key-length"

	FlagVerbose = "verbose"
	FlagRuns    = "probe-runs"
)

var resultColor = color.New(color.FgGreen)

func newCalibrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calibrate [<desired-duration>]",
		Args:  cobra.ExactArgs(1),
		Short: "Calibrate the values for Argon2 so the hashing operation takes the desired time.",
		Long: `Calibrate the configuration values for Argon2 by probing the execution time.
Note that the values depend on the machine you run the hashing on.
When choosing the desired time, UX is in conflict with security. Security should really win out here, therefore we recommend 1s.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			desiredDuration, err := time.ParseDuration(args[0])
			if err != nil {
				return err
			}
			verbose := flagx.MustGetBool(cmd, FlagVerbose)
			runs := flagx.MustGetInt(cmd, FlagRuns)

			config, err := configFromFlags(cmd)
			if err != nil {
				return err
			}
			hasher := hash.NewHasherArgon2(config)

			var maxMemory, adjustMemory datasize.ByteSize
			if err := maxMemory.UnmarshalText([]byte(flagx.MustGetString(cmd, FlagMaxMemory))); err != nil {
				return err
			}
			if err := adjustMemory.UnmarshalText([]byte(flagx.MustGetString(cmd, FlagAdjustMemory))); err != nil {
				return err
			}

			var currentDuration time.Duration

			if verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "Increasing memory to get over %s:\n", desiredDuration)
			}
			for {
				if maxMemory != 0 && config.c.Memory > toKB(maxMemory) {
					// don't further increase memory
					if verbose {
						fmt.Fprintln(cmd.ErrOrStderr(), "  ouch, hit the memory limit there")
					}
					config.c.Memory = toKB(maxMemory)
					break
				}

				currentDuration, err = probe(cmd, hasher, runs, verbose)
				if err != nil {
					return err
				}
				if verbose {
					fmt.Fprintf(cmd.ErrOrStderr(), "  took %s with %s of memory\n", currentDuration, config.getMemFormat())
				}

				if currentDuration > desiredDuration {
					if config.c.Memory <= toKB(adjustMemory) {
						// adjusting the memory would now result in <= 0B
						adjustMemory = adjustMemory >> 1
					}
					config.c.Memory -= toKB(adjustMemory)
					break
				}

				// adjust config
				config.c.Memory += toKB(adjustMemory)
			}

			if verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "Decreasing memory to get under %s:\n", desiredDuration)
			}
			for {
				currentDuration, err = probe(cmd, hasher, runs, verbose)
				if err != nil {
					return err
				}
				if verbose {
					fmt.Fprintf(cmd.ErrOrStderr(), "  took %s with %s of memory\n", currentDuration, config.getMemFormat())
				}

				if currentDuration < desiredDuration {
					break
				}
				if config.c.Memory <= toKB(adjustMemory) {
					// adjusting the memory would now result in <= 0B
					adjustMemory = adjustMemory >> 1
				}
				// adjust config
				config.c.Memory -= toKB(adjustMemory)
			}
			if verbose {
				_, _ = resultColor.Fprintf(cmd.ErrOrStderr(), "Settled on %s of memory.\n", config.getMemFormat())
				fmt.Fprintf(cmd.ErrOrStderr(), "Increasing iterations to get over %s:\n", desiredDuration)
			}
			for {
				currentDuration, err = probe(cmd, hasher, runs, verbose)
				if err != nil {
					return err
				}
				if verbose {
					fmt.Fprintf(cmd.ErrOrStderr(), "  took %s with %d iterations\n", currentDuration, config.c.Iterations)
				}

				if currentDuration > desiredDuration {
					config.c.Iterations -= 1
					break
				}
				// adjust config
				config.c.Iterations += 1
			}

			if verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "Decreasing iterations to get under %s:\n", desiredDuration)
			}
			for {
				currentDuration, err = probe(cmd, hasher, runs, verbose)
				if err != nil {
					return err
				}
				if verbose {
					fmt.Fprintf(cmd.ErrOrStderr(), "  took %s with %d iterations\n", currentDuration, config.c.Iterations)
				}

				// break also when iterations is 1; this catches the case where 1 was only slightly under the desired time and took longer a bit longer on another run
				if currentDuration < desiredDuration || config.c.Iterations == 1 {
					break
				}
				// adjust config
				config.c.Iterations -= 1
			}
			if verbose {
				_, _ = resultColor.Fprintf(cmd.ErrOrStderr(), "Settled on %d iterations.\n\n", config.c.Iterations)
			}

			e := json.NewEncoder(cmd.OutOrStdout())
			e.SetIndent("", "  ")
			return e.Encode(config.c)
		},
	}

	registerArgon2Flags(cmd.Flags())
	cmd.Flags().BoolP(FlagVerbose, "v", false, "verbose output")
	cmd.Flags().IntP(FlagRuns, "r", 2, "runs per probe, median of all runs is taken as the result")

	return cmd
}

func toKB(b datasize.ByteSize) uint32 {
	return uint32(b / datasize.KB)
}

func probe(cmd *cobra.Command, hasher hash.Hasher, runs int, verbose bool) (time.Duration, error) {
	start := time.Now()

	var mid time.Time
	for i := 0; i < runs; i++ {
		mid = time.Now()
		_, err := hasher.Generate([]byte("password"))
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Could not generate a hash: %s\n", err)
			return 0, cmdx.FailSilently(cmd)
		}
		if verbose {
			fmt.Fprintf(cmd.OutOrStdout(), "    took %s in try %d\n", time.Since(mid), i)
		}
	}

	return time.Duration(int64(time.Since(start)) / int64(runs)), nil
}

func registerArgon2Flags(flags *pflag.FlagSet) {
	flags.StringP(FlagStartMemory, "m", "4GB", "amount of memory to start probing at")
	flags.String(FlagMaxMemory, "", "maximum memory allowed (default no limit)")
	flags.String(FlagAdjustMemory, "1GB", "amount by which the memory is adjusted in every step while probing")

	flags.Uint32P(FlagStartIterations, "i", 1, "number of iterations to start probing at")

	flags.Uint8(FlagParallelism, configuration.Argon2DefaultParallelism, "number of threads to use")

	flags.Uint32(FlagSaltLength, configuration.Argon2DefaultSaltLength, "length of the salt in bytes")

	flags.Uint32(FlagKeyLength, configuration.Argon2DefaultKeyLength, "length of the key in bytes")
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
	var mem datasize.ByteSize
	if err := mem.UnmarshalText([]byte(flagx.MustGetString(cmd, FlagStartMemory))); err != nil {
		return nil, err
	}

	return &argon2Config{
		c: configuration.HasherArgon2Config{
			Parallelism: getUint8(FlagParallelism),
			Memory:      uint32(mem / datasize.KB),
			Iterations:  getUint32(FlagStartIterations),
			SaltLength:  getUint32(FlagSaltLength),
			KeyLength:   getUint32(FlagKeyLength),
		},
	}, err
}
