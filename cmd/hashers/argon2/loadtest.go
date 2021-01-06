package argon2

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/inhies/go-bytesize"
	"github.com/montanaflynn/stats"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/ory/kratos/hash"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/configx"
)

type resultTable struct {
	Total    time.Duration       `json:"total"`
	Median   time.Duration       `json:"median"`
	StdDev   time.Duration       `json:"std_deviation"`
	Min      time.Duration       `json:"min"`
	Max      time.Duration       `json:"max"`
	MemStats []*runtime.MemStats `json:"-"`
}

var _ cmdx.OutputEntry = &resultTable{}

func (r *resultTable) Header() []string {
	return []string{"TOTAL SAMPLE TIME", "MEDIAN", "STANDARD DEVIATION", "MIN", "MAX", "MEMORY REQUIRED"}
}

func (r *resultTable) Fields() []string {
	var sysMemMax uint64
	for i := range r.MemStats {
		if sysMemMax < r.MemStats[i].Sys {
			sysMemMax = r.MemStats[i].Sys
		}
	}

	return []string{
		r.Total.String(),
		r.Median.String(),
		r.StdDev.String(),
		r.Min.String(),
		r.Max.String(),
		bytesize.ByteSize(sysMemMax).String(),
	}
}

func (r *resultTable) Interface() interface{} {
	return r
}

func newLoadTestCmd() *cobra.Command {
	flagConf := &argon2Config{}

	cmd := &cobra.Command{
		Use:  "load-test <ops-per-minute>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			perMinute, err := strconv.ParseInt(args[0], 0, 0)
			if err != nil {
				return err
			}

			conf, err := configProvider(cmd, flagConf)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.ErrOrStderr(), "This command takes about 1 minute to collect all necessary data, be patient.")

			hasher := hash.NewHasherArgon2(conf)

			allDone := make(chan struct{})

			var memStats []*runtime.MemStats
			go func() {
				clock := time.NewTicker(time.Second)
				defer func() {
					clock.Stop()
				}()

				for {
					select {
					case <-cmd.Context().Done():
						return
					case <-allDone:
						return
					case <-clock.C:
						ms := runtime.MemStats{}
						runtime.ReadMemStats(&ms)
						memStats = append(memStats, &ms)
					}
				}
			}()

			go func() {
				input := make([]byte, 1)
				for {
					n, err := cmd.InOrStdin().Read(input)
					if err != nil {
						return
					}
					if n != 0 {
						_, _ = color.New(color.FgRed).Fprintln(cmd.ErrOrStderr(), "I SAID BE PATIENT!!!")
						return
					}

					select {
					case <-allDone:
						return
					case <-cmd.Context().Done():
						return
					case <-time.After(time.Millisecond):
					}
				}
			}()

			calcTimes := make([]time.Duration, perMinute)
			eg, _ := errgroup.WithContext(cmd.Context())

			startAll := time.Now()

			for i := 0; i < int(perMinute); i++ {
				eg.Go(func(i int) func() error {
					return func() error {
						// wait randomly before starting, max. 1 minute
						// #nosec G404 - just a timeout to collect statistical data
						time.Sleep(time.Duration(rand.Intn(int(time.Minute))))

						start := time.Now()
						_, err := hasher.Generate([]byte("password"))
						if err != nil {
							return err
						}

						calcTimes[i] = time.Since(start)
						return nil
					}
				}(i))
			}

			if err := eg.Wait(); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error during hashing: %+v\n", err)
				return cmdx.FailSilently(cmd)
			}
			totalTime := time.Since(startAll)
			close(allDone)

			calcData := stats.LoadRawData(calcTimes)

			duration := func(f func() (float64, error)) time.Duration {
				v, err := f()
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Unexpected maths error: %+v\nRaw Data: %+v\n", err, calcTimes)
				}
				return time.Duration(int64(v))
			}

			cmdx.PrintRow(cmd, &resultTable{
				Total:    totalTime,
				Median:   duration(calcData.Mean),
				StdDev:   duration(calcData.StandardDeviation),
				Min:      duration(calcData.Min),
				Max:      duration(calcData.Max),
				MemStats: memStats,
			})
			return nil
		},
	}

	registerArgon2ConfigFlags(cmd.Flags(), flagConf)
	configx.RegisterFlags(cmd.Flags())
	cmdx.RegisterFormatFlags(cmd.Flags())

	return cmd
}
