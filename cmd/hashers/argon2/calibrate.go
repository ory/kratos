package argon2

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/inhies/go-bytesize"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/hash"
	"github.com/ory/x/cmdx"
	"github.com/spf13/cobra"
)

type (
	argon2Config struct {
		c configuration.HasherArgon2Config
	}
	localByteSize bytesize.ByteSize
)

func (c *argon2Config) HasherArgon2() *configuration.HasherArgon2Config {
	return &c.c
}

func (c *argon2Config) getMemFormat() string {
	return (bytesize.ByteSize(c.c.Memory) * bytesize.KB).String()
}

func (b *localByteSize) Type() string {
	return "byte_size"
}

func (b *localByteSize) Set(v string) error {
	return (*bytesize.ByteSize)(b).UnmarshalText([]byte(v))
}

func (b *localByteSize) String() string {
	return (*bytesize.ByteSize)(b).String()
}

const (
	FlagStartMemory     = "start-memory"
	FlagMaxMemory       = "max-memory"
	FlagAdjustMemory    = "adjust-memory-by"
	FlagStartIterations = "start-iterations"
	FlagParallelism     = "parallelism"
	FlagSaltLength      = "salt-length"
	FlagKeyLength       = "key-length"

	FlagQuiet = "quiet"
	FlagRuns  = "probe-runs"
)

var resultColor = color.New(color.FgGreen)

func newCalibrateCmd() *cobra.Command {
	var (
		maxMemory, adjustMemory, startMemory localByteSize = 0, localByteSize(bytesize.GB), localByteSize(4 * bytesize.GB)
		quiet bool
		runs int
	)

	config := &argon2Config{
		c: configuration.HasherArgon2Config{},
	}

	cmd := &cobra.Command{
		Use:   "calibrate [<desired-duration>]",
		Args:  cobra.ExactArgs(1),
		Short: "Computes Optimal Argon2 Parameters.",
		Long: `This command helps you calibrate the configuration parameters for Argon2. Password hashing is a trade-off between security, resource consumption, and user experience. Resource consumption should not be too high and the login should not take too long.

We recommend that the login process takes between half a second and one second for password hashing, giving a good balance between security and user experience.

Please note that the values depend on the machine you run the hashing on. If you have RAM constraints please choose lower memory targets to avoid out of memory panics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			desiredDuration, err := time.ParseDuration(args[0])
			if err != nil {
				return err
			}

			config.c.Memory = toKB(startMemory)

			hasher := hash.NewHasherArgon2(config)

			var currentDuration time.Duration

			if !quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "Increasing memory to get over %s:\n", desiredDuration)
			}

			for {
				if maxMemory != 0 && config.c.Memory > toKB(maxMemory) {
					// don't further increase memory
					if !quiet {
						fmt.Fprintln(cmd.ErrOrStderr(), "  ouch, hit the memory limit there")
					}
					config.c.Memory = toKB(maxMemory)
					break
				}

				currentDuration, err = probe(cmd, hasher, runs, quiet)
				if err != nil {
					return err
				}

				if !quiet {
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

			if !quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "Decreasing memory to get under %s:\n", desiredDuration)
			}

			for {
				currentDuration, err = probe(cmd, hasher, runs, quiet)
				if err != nil {
					return err
				}

				if !quiet {
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

			if !quiet {
				_, _ = resultColor.Fprintf(cmd.ErrOrStderr(), "Settled on %s of memory.\n", config.getMemFormat())
				fmt.Fprintf(cmd.ErrOrStderr(), "Increasing iterations to get over %s:\n", desiredDuration)
			}

			for {
				currentDuration, err = probe(cmd, hasher, runs, quiet)
				if err != nil {
					return err
				}

				if !quiet {
					fmt.Fprintf(cmd.ErrOrStderr(), "  took %s with %d iterations\n", currentDuration, config.c.Iterations)
				}

				if currentDuration > desiredDuration {
					config.c.Iterations -= 1
					break
				}

				// adjust config
				config.c.Iterations += 1
			}

			if !quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "Decreasing iterations to get under %s:\n", desiredDuration)
			}

			for {
				currentDuration, err = probe(cmd, hasher, runs, quiet)
				if err != nil {
					return err
				}

				if !quiet {
					fmt.Fprintf(cmd.ErrOrStderr(), "  took %s with %d iterations\n", currentDuration, config.c.Iterations)
				}

				// break also when iterations is 1; this catches the case where 1 was only slightly under the desired time and took longer a bit longer on another run
				if currentDuration < desiredDuration || config.c.Iterations == 1 {
					break
				}

				// adjust config
				config.c.Iterations -= 1
			}
			if !quiet {
				_, _ = resultColor.Fprintf(cmd.ErrOrStderr(), "Settled on %d iterations.\n\n", config.c.Iterations)
			}

			e := json.NewEncoder(cmd.OutOrStdout())
			e.SetIndent("", "  ")
			return e.Encode(config.c)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&quiet, FlagQuiet, "q", false, "Quiet output.")
	flags.IntVarP(&runs, FlagRuns, "r", 2, "Runs per probe, median of all runs is taken as the result.")

	flags.VarP(&startMemory, FlagStartMemory, "m", "Amount of memory to start probing at.")
	flags.Var(&maxMemory, FlagMaxMemory, "Maximum memory allowed (default no limit).")
	flags.Var(&adjustMemory, FlagAdjustMemory, "Amount by which the memory is adjusted in every step while probing.")

	flags.Uint32VarP(&config.c.Iterations, FlagStartIterations, "i", 1, "Number of iterations to start probing at.")

	flags.Uint8Var(&config.c.Parallelism, FlagParallelism, configuration.Argon2DefaultParallelism, "Number of threads to use.")

	flags.Uint32Var(&config.c.SaltLength, FlagSaltLength, configuration.Argon2DefaultSaltLength, "Length of the salt in bytes.")
	flags.Uint32Var(&config.c.KeyLength, FlagKeyLength, configuration.Argon2DefaultKeyLength, "Length of the key in bytes.")

	return cmd
}

func toKB(b localByteSize) uint32 {
	return uint32(b / localByteSize(bytesize.KB))
}

func probe(cmd *cobra.Command, hasher hash.Hasher, runs int, quiet bool) (time.Duration, error) {
	start := time.Now()

	var mid time.Time
	for i := 0; i < runs; i++ {
		mid = time.Now()
		_, err := hasher.Generate([]byte("password"))
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Could not generate a hash: %s\n", err)
			return 0, cmdx.FailSilently(cmd)
		}
		if !quiet {
			fmt.Fprintf(cmd.OutOrStdout(), "    took %s in try %d\n", time.Since(mid), i)
		}
	}

	return time.Duration(int64(time.Since(start)) / int64(runs)), nil
}
