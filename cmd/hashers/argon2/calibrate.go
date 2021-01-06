package argon2

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ory/x/cmdx"

	"github.com/fatih/color"
	"github.com/inhies/go-bytesize"
	"github.com/spf13/cobra"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hash"
)

const (
	FlagStartMemory     = "start-memory"
	FlagMaxMemory       = "max-memory"
	FlagAdjustMemory    = "adjust-memory-by"
	FlagStartIterations = "start-iterations"
	FlagMaxConcurrent   = "max-concurrent"

	FlagQuiet = "quiet"
	FlagRuns  = "probe-runs"
)

var resultColor = color.New(color.FgGreen)

func newCalibrateCmd() *cobra.Command {
	var (
		maxMemory, adjustMemory bytesize.ByteSize = 0, 1 * bytesize.GB
		quiet                   bool
		runs                    int
	)

	aconfig := &argon2Config{
		c: config.HasherArgon2Config{},
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

			if aconfig.memory == 0 {
				aconfig.memory = 4 * bytesize.GB
			}

			hasher := hash.NewHasherArgon2(aconfig)

			var currentDuration time.Duration

			if !quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "Increasing memory to get over %s:\n", desiredDuration)
			}

			for {
				if maxMemory != 0 && aconfig.memory > maxMemory {
					// don't further increase memory
					if !quiet {
						fmt.Fprintln(cmd.ErrOrStderr(), "  ouch, hit the memory limit there")
					}
					aconfig.memory = maxMemory
					break
				}

				currentDuration, err = probe(cmd, hasher, runs, quiet)
				if err != nil {
					return err
				}

				if !quiet {
					fmt.Fprintf(cmd.ErrOrStderr(), "  took %s with %s of memory\n", currentDuration, aconfig.getMemFormat())
				}

				if currentDuration > desiredDuration {
					if aconfig.memory <= adjustMemory {
						// adjusting the memory would now result in <= 0B
						adjustMemory = adjustMemory >> 1
					}
					aconfig.memory -= adjustMemory
					break
				}

				// adjust config
				aconfig.memory += adjustMemory
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
					fmt.Fprintf(cmd.ErrOrStderr(), "  took %s with %s of memory\n", currentDuration, aconfig.getMemFormat())
				}

				if currentDuration < desiredDuration {
					break
				}

				if aconfig.memory <= adjustMemory {
					// adjusting the memory would now result in <= 0B
					adjustMemory = adjustMemory >> 1
				}

				// adjust config
				aconfig.memory -= adjustMemory
			}

			if !quiet {
				_, _ = resultColor.Fprintf(cmd.ErrOrStderr(), "Settled on %s of memory.\n", aconfig.getMemFormat())
				fmt.Fprintf(cmd.ErrOrStderr(), "Increasing iterations to get over %s:\n", desiredDuration)
			}

			for {
				currentDuration, err = probe(cmd, hasher, runs, quiet)
				if err != nil {
					return err
				}

				if !quiet {
					fmt.Fprintf(cmd.ErrOrStderr(), "  took %s with %d iterations\n", currentDuration, aconfig.c.Iterations)
				}

				if currentDuration > desiredDuration {
					aconfig.c.Iterations -= 1
					break
				}

				// adjust config
				aconfig.c.Iterations += 1
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
					fmt.Fprintf(cmd.ErrOrStderr(), "  took %s with %d iterations\n", currentDuration, aconfig.c.Iterations)
				}

				// break also when iterations is 1; this catches the case where 1 was only slightly under the desired time and took longer a bit longer on another run
				if currentDuration < desiredDuration || aconfig.c.Iterations == 1 {
					break
				}

				// adjust config
				aconfig.c.Iterations -= 1
			}
			if !quiet {
				_, _ = resultColor.Fprintf(cmd.ErrOrStderr(), "Settled on %d iterations.\n\n", aconfig.c.Iterations)
			}

			e := json.NewEncoder(cmd.OutOrStdout())
			e.SetIndent("", "  ")
			return e.Encode(aconfig.c)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&quiet, FlagQuiet, "q", false, "Quiet output.")
	flags.IntVarP(&runs, FlagRuns, "r", 2, "Runs per probe, median of all runs is taken as the result.")

	flags.VarP(&aconfig.memory, FlagStartMemory, "m", "Amount of memory to start probing at.")
	flags.Var(&maxMemory, FlagMaxMemory, "Maximum memory allowed (default no limit).")
	flags.Var(&adjustMemory, FlagAdjustMemory, "Amount by which the memory is adjusted in every step while probing.")

	flags.Uint32VarP(&aconfig.c.Iterations, FlagStartIterations, "i", 1, "Number of iterations to start probing at.")

	flags.Uint8(FlagMaxConcurrent, 16, "Maximum number of concurrent hashing operations.")

	registerArgon2ConstantConfigFlags(flags, aconfig)

	return cmd
}

func probe(cmd *cobra.Command, hasher hash.Hasher, runs int, quiet bool) (time.Duration, error) {
	start := time.Now()
	//concurrent := flagx.MustGetUint8(cmd, FlagMaxConcurrent)

	var mid time.Time
	for i := 0; i < runs; i++ {
		//errs := make(chan error, concurrent)

		mid = time.Now()
		//for j := 0; j < concurrent; j-- {
		//	go func() {
		_, err := hasher.Generate([]byte("password"))
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Could not generate a hash: %s\n", err)
			return 0, cmdx.FailSilently(cmd)
		}
		//	}()
		//
		//}

		if !quiet {
			fmt.Fprintf(cmd.OutOrStdout(), "    took %s in try %d\n", time.Since(mid), i)
		}
	}

	return time.Duration(int64(time.Since(start)) / int64(runs)), nil
}
