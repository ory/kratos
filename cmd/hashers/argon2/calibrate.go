package argon2

import (
	"fmt"
	"github.com/ory/x/configx"
	"io"
	"runtime"
	"strconv"
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

	FlagRuns = "probe-runs"
)

type (
	colorWriter struct {
		c *color.Color
		w io.Writer
	}
	loadResult struct {
		r *resultTable
		v *config.HasherArgon2Config
	}
	loadResults []*loadResult
)

var _ cmdx.Table = loadResults{}

func (l loadResults) Header() []string {
	return append((&resultTable{}).Header(), "MEMORY PARAM", "ITERATIONS PARAM")
}

func (l loadResults) Table() [][]string {
	t := make([][]string, len(l))

	for i, r := range l {
		t[i] = append(r.r.Columns(), fmt.Sprintf("%d", r.v.Memory), fmt.Sprintf("%d", r.v.Iterations))
	}

	return t
}

func (l loadResults) Interface() interface{} {
	return l
}

func (l loadResults) Len() int {
	return len(l)
}

func (c *colorWriter) Write(o []byte) (int, error) {
	return c.c.Fprintf(c.w, "%s", o)
}

func newCalibrateCmd() *cobra.Command {
	var (
		maxMemory, adjustMemory bytesize.ByteSize = 0, 1 * bytesize.GB
		runs                    int
	)

	flagConfig := &argon2Config{
		c: config.HasherArgon2Config{},
	}

	cmd := &cobra.Command{
		Use:   "calibrate <requests-per-minute>",
		Args:  cobra.ExactArgs(1),
		Short: "Computes Optimal Argon2 Parameters.",
		Long: `This command helps you calibrate the configuration parameters for Argon2. Password hashing is a trade-off between security, resource consumption, and user experience. Resource consumption should not be too high and the login should not take too long.

We recommend that the login process takes between half a second and one second for password hashing, giving a good balance between security and user experience.

Please note that the values depend on the machine you run the hashing on. If you have RAM constraints please choose lower memory targets to avoid out of memory panics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			progressPrinter := cmdx.NewLoudErrPrinter(cmd)
			resultPrinter := cmdx.NewLoudPrinter(cmd, &colorWriter{c: color.New(color.FgGreen), w: cmd.ErrOrStderr()})

			conf, err := configProvider(cmd, flagConfig)
			if err != nil {
				return err
			}

			// always take start flags, or their default
			conf.memory = flagConfig.memory
			conf.c.Iterations = flagConfig.c.Iterations

			desiredDuration := conf.c.MinimalDuration

			reqPerMin, err := strconv.ParseInt(args[0], 0, 0)
			if err != nil {
				// we want the error and usage string printed so just return
				return err
			}

			if conf.memory == 0 {
				conf.memory = 4 * bytesize.GB
			}

			hasher := hash.NewHasherArgon2(conf)

			var currentDuration time.Duration

			_, _ = progressPrinter.Printf("Increasing memory to get over %s:\n", desiredDuration)

			for {
				if maxMemory != 0 && conf.memory > maxMemory {
					// don't further increase memory
					_, _ = progressPrinter.Println("  ouch, hit the memory limit there")
					conf.memory = maxMemory
					break
				}

				currentDuration, err = probe(cmd, hasher, runs, progressPrinter)
				if err != nil {
					return err
				}

				_, _ = progressPrinter.Printf("  took %s with %s of memory\n", currentDuration, conf.getMemFormat())

				if currentDuration > desiredDuration {
					for conf.memory <= adjustMemory {
						// adjusting the memory would now result in <= 0B
						adjustMemory = adjustMemory >> 1
					}
					conf.memory -= adjustMemory
					break
				}

				// adjust config
				conf.memory += adjustMemory
			}

			_, _ = progressPrinter.Printf("Decreasing memory to get under %s:\n", desiredDuration)

			for {
				currentDuration, err = probe(cmd, hasher, runs, progressPrinter)
				if err != nil {
					return err
				}

				_, _ = progressPrinter.Printf("  took %s with %s of memory\n", currentDuration, conf.getMemFormat())

				if currentDuration < desiredDuration {
					break
				}

				for conf.memory <= adjustMemory {
					// adjusting the memory would now result in <= 0B
					adjustMemory = adjustMemory >> 1
				}

				// adjust config
				conf.memory -= adjustMemory
			}

			_, _ = resultPrinter.Printf("Settled on %s of memory.\n", conf.getMemFormat())
			_, _ = progressPrinter.Printf("Increasing iterations to get over %s:\n", desiredDuration)

			for {
				currentDuration, err = probe(cmd, hasher, runs, progressPrinter)
				if err != nil {
					return err
				}

				_, _ = progressPrinter.Printf("  took %s with %d iterations\n", currentDuration, conf.c.Iterations)

				if currentDuration > desiredDuration {
					conf.c.Iterations -= 1
					break
				}

				// adjust config
				conf.c.Iterations += 1
			}

			_, _ = progressPrinter.Printf("Decreasing iterations to get under %s:\n", desiredDuration)

			for {
				currentDuration, err = probe(cmd, hasher, runs, progressPrinter)
				if err != nil {
					return err
				}

				_, _ = progressPrinter.Printf("  took %s with %d iterations\n", currentDuration, conf.c.Iterations)

				// break also when iterations is 1; this catches the case where 1 was only slightly under the desired time and took longer a bit longer on another run
				if currentDuration < desiredDuration || conf.c.Iterations == 1 {
					break
				}

				// adjust config
				conf.c.Iterations -= 1
			}
			_, _ = resultPrinter.Printf("Settled on %d iterations.\n", conf.c.Iterations)

			results := make(loadResults, 5)
			for i := 0; i < len(results); i++ {
				_, _ = progressPrinter.Printf("\nRunning load test to simulate %d login requests per minute.\n", reqPerMin)

				res, err := runLoadTest(cmd, conf, int(reqPerMin))
				if err != nil {
					return err
				}
				cv, _ := conf.HasherArgon2()
				ccv := *cv
				results[i] = &loadResult{r: res, v: &ccv}

				_, _ = progressPrinter.Println("The load test result is:\n")
				cmdx.PrintRow(cmd, res)

				switch {
				// too fast
				case res.Median < conf.c.MinimalDuration:
					_, _ = progressPrinter.Printf("The median was %s under the minimal duration of %s, going to increase the hash cost.\n", conf.c.MinimalDuration-res.Median, conf.c.MinimalDuration)

					// try to increase memory first
					if res.MaxMem+64*bytesize.MB < maxMemory {
						// only small amounts of memory are sensible as we already benchmarked a single, non-concurrent request
						conf.memory += 64 * bytesize.MB
						_, _ = progressPrinter.Printf("Increasing memory to %s\n", conf.memory)
					} else {
						// increasing memory is not allowed by maxMemory, therefore increase CPU load
						conf.c.Iterations++
						_, _ = progressPrinter.Printf("Increasing iterations to %d\n", conf.c.Iterations)
					}
				// too much memory
				case res.MaxMem > conf.c.DedicatedMemory:
					_, _ = progressPrinter.Printf("The required memory was %s more than the maximum allowed of %s.\n", res.MaxMem-maxMemory, conf.c.DedicatedMemory)

					conf.memory -= (res.MaxMem - conf.c.DedicatedMemory) / bytesize.ByteSize(reqPerMin)
					_, _ = progressPrinter.Printf("Decreasing memory to %s\n", conf.memory)
				// too slow
				case res.Max > conf.c.ExpectedDeviation+conf.c.MinimalDuration:
					_, _ = progressPrinter.Printf("The longest request took %s longer than the longest acceptable time of %s, going to decrease the hash cost.\n", res.Max-conf.c.ExpectedDeviation+conf.c.MinimalDuration, conf.c.ExpectedDeviation+conf.c.MinimalDuration)

					// try to decrease iterations first
					if conf.c.Iterations > 1 {
						conf.c.Iterations--
						_, _ = progressPrinter.Printf("Decreasing iterations to %d\n", conf.c.Iterations)
					} else {
						// decreasing iterations is not possible anymore, decreasing memory
						// only small amounts of memory are sensible as we already benchmarked a single, non-concurrent request
						conf.memory -= 64 * bytesize.MB
						_, _ = progressPrinter.Printf("Decreasing memory to %s\n", conf.memory)
					}
				// too high deviation
				case res.StdDev > conf.c.ExpectedDeviation:
					_, _ = progressPrinter.Printf("The deviation was %s more than the expected deviation of %s.\n", res.StdDev-conf.c.ExpectedDeviation, conf.c.ExpectedDeviation)

					// try to decrease iterations first
					if conf.c.Iterations > 1 {
						conf.c.Iterations--
						_, _ = progressPrinter.Printf("Decreasing iterations to %d\n", conf.c.Iterations)
					} else {
						// decreasing iterations is not possible anymore, decreasing memory
						// only small amounts of memory are sensible as we already benchmarked a single, non-concurrent request
						conf.memory -= 64 * bytesize.MB
						_, _ = progressPrinter.Printf("Decreasing memory to %s\n", conf.memory)
					}
				// all values seem reasonable
				default:
					_, _ = progressPrinter.Println("These values look good to me.\n")
					cmdx.PrintRow(cmd, conf)
					return nil
				}
			}

			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Could not automatically determine good parameters. Have a look at all the measurements taken and select acceptable values yourself. Have a look in the docs for more information: https://www.ory.sh/kratos/docs/debug/performance-out-of-memory-password-hashing-argon2")
			cmdx.PrintTable(cmd, results)
			return nil
		},
	}

	flags := cmd.Flags()

	flags.IntVarP(&runs, FlagRuns, "r", 2, "Runs per probe, median of all runs is taken as the result.")

	flags.VarP(&flagConfig.memory, FlagStartMemory, "m", "Amount of memory to start probing at.")
	flags.Var(&maxMemory, FlagMaxMemory, "Maximum memory allowed (default no limit).")
	flags.Var(&adjustMemory, FlagAdjustMemory, "Amount by which the memory is adjusted in every step while probing.")

	flags.Uint32VarP(&flagConfig.c.Iterations, FlagStartIterations, "i", 1, "Number of iterations to start probing at.")

	flags.Uint8(FlagMaxConcurrent, 16, "Maximum number of concurrent hashing operations.")

	registerArgon2ConstantConfigFlags(flags, flagConfig)
	cmdx.RegisterFormatFlags(flags)
	configx.RegisterFlags(flags)

	return cmd
}

func probe(cmd *cobra.Command, hasher hash.Hasher, runs int, progressPrinter *cmdx.ConditionalPrinter) (time.Duration, error) {
	// force GC at the start of the experiment
	runtime.GC()

	start := time.Now()

	var mid time.Time
	for i := 0; i < runs; i++ {
		mid = time.Now()
		_, err := hasher.Generate([]byte("password"))
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Could not generate a hash: %s\n", err)
			return 0, cmdx.FailSilently(cmd)
		}

		_, _ = progressPrinter.Printf("    took %s in try %d\n", time.Since(mid), i)
	}
	return time.Duration(int64(time.Since(start)) / int64(runs)), nil
}
