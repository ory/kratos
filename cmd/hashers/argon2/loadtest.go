package argon2

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/ory/x/flagx"

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
	TotalTime  time.Duration     `json:"total_time"`
	MedianTime time.Duration     `json:"median_request_time"`
	StdDev     time.Duration     `json:"std_deviation"`
	MinTime    time.Duration     `json:"min_request_time"`
	MaxTime    time.Duration     `json:"max_request_time"`
	MaxMem     bytesize.ByteSize `json:"mem_used"`
}

var (
	ErrSampleTimeExceeded        = fmt.Errorf("the sample time was exceeded")
	ErrMemoryConsumptionExceeded = fmt.Errorf("the memory consumption was exceeded")

	_ cmdx.TableRow = &resultTable{}
)

func (r *resultTable) Header() []string {
	return []string{"TOTAL SAMPLE TIME", "MEDIAN REQUEST TIME", "STANDARD DEVIATION", "MIN REQUEST TIME", "MAX REQUEST TIME", "MEMOry USED"}
}

func (r *resultTable) Columns() []string {
	return []string{
		r.TotalTime.String(),
		r.MedianTime.String(),
		r.StdDev.String(),
		r.MinTime.String(),
		r.MaxTime.String(),
		r.MaxMem.String(),
	}
}

func (r *resultTable) Interface() interface{} {
	return r
}

func newLoadTestCmd() *cobra.Command {
	flagConf := &argon2Config{}

	cmd := &cobra.Command{
		Use:   "load-test <authentication-requests-per-minute>",
		Short: "Simulate the password hashing with a number of concurrent requests/minute.",
		Long:  "Simulates a number of concurrent authentication requests per minute. Gives statistical data about the measured performance and resource consumption. Can be used to tune and test the hashing parameters for peak demand situations.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			perMinute, err := strconv.ParseInt(args[0], 0, 0)
			if err != nil {
				return err
			}

			conf, err := configProvider(cmd, flagConf)
			if err != nil {
				return err
			}

			if !flagx.MustGetBool(cmd, cmdx.FlagQuiet) {
				fmt.Fprintln(cmd.ErrOrStderr(), "The hashing configuration used is:")
				cmdx.PrintRow(cmd, conf)
			}

			res, err := runLoadTest(cmd, conf, int(perMinute))
			if err != nil {
				return err
			}

			cmdx.PrintRow(cmd, res)
			return nil
		},
	}

	registerArgon2ConfigFlags(cmd.Flags(), flagConf)
	configx.RegisterFlags(cmd.Flags())
	cmdx.RegisterFormatFlags(cmd.Flags())

	return cmd
}

func runLoadTest(cmd *cobra.Command, conf *argon2Config, reqPerMin int) (*resultTable, error) {
	// force GC at the start of the experiment
	runtime.GC()

	sampleTime := time.Minute / 3
	reqNum := reqPerMin / int(time.Minute/sampleTime)

	progressPrinter := cmdx.NewLoudErrPrinter(cmd)
	_, _ = progressPrinter.Printf("It takes about %s to collect all necessary data, please be patient.\n", sampleTime)

	ctx, cancel := context.WithCancel(cmd.Context())
	hasher := hash.NewHasherArgon2(conf)
	allDone := make(chan struct{})
	startAll := time.Now()
	var cancelReason error

	var memStats []uint64
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
				// cancel if the allowed time is exceeded by 110%
				if time.Since(startAll) > ((sampleTime+conf.localConfig.ExpectedDuration+conf.localConfig.ExpectedDeviation)/100)*110 {
					cancelReason = ErrSampleTimeExceeded
					cancel()
					return
				}

				ms := runtime.MemStats{}
				runtime.ReadMemStats(&ms)

				// cancel if memory is exceeded by 110%
				if ms.HeapAlloc > (uint64(conf.localConfig.DedicatedMemory)/100)*110 {
					cancelReason = ErrMemoryConsumptionExceeded
					cancel()
					return
				}

				memStats = append(memStats, ms.HeapAlloc)
			}
		}
	}()

	go func() {
		// don't read std_in when quiet
		if flagx.MustGetBool(cmd, cmdx.FlagQuiet) {
			return
		}

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

	calcTimes := make([]time.Duration, reqNum)
	eg, _ := errgroup.WithContext(ctx)

	for i := 0; i < reqNum; i++ {
		eg.Go(func(i int) func() error {
			return func() error {
				// wait randomly before starting, max. sample time
				// #nosec G404 - just a timeout to collect statistical data
				t := time.Duration(rand.Intn(int(sampleTime)))
				timer := time.NewTimer(t)
				defer timer.Stop()

				select {
				case <-ctx.Done():
					return nil
				case <-timer.C:
				}

				start := time.Now()
				done := make(chan struct{})
				var err error

				go func() {
					_, err = hasher.Generate(ctx, []byte("password"))
					close(done)
				}()

				select {
				case <-ctx.Done():
					return nil
				case <-done:
					if err != nil {
						return err
					}

					calcTimes[i] = time.Since(start)
					return nil
				}
			}
		}(i))
	}

	if err := eg.Wait(); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error during hashing: %+v\n", err)
		return nil, cmdx.FailSilently(cmd)
	}
	switch cancelReason {
	case ErrSampleTimeExceeded:
		memUsed, err2 := stats.LoadRawData(memStats).Max()
		if err2 != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Unexpected maths error: %+v\nRaw Data: %+v\n", cancelReason, memStats)
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "The hashing load test took too long. This indicates that you don't have enough resources to handle %d login requests per minute with the desired minimal time of %s. The memory used was %s. Either dedicate more CPU/memory, or decrease the hashing cost (memory and iterations parameters).\n", reqPerMin, conf.localConfig.ExpectedDuration, bytesize.ByteSize(memUsed))
		return nil, cmdx.FailSilently(cmd)
	case ErrMemoryConsumptionExceeded:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "The hashing load test exceeded the memory limit of %s. This indicates that you don't have enough resources to handle %d login requests per minute with the desired minimal time of %s. Either dedicate more memory, or decrease the hashing cost (memory and iterations parameters).\n", conf.localConfig.DedicatedMemory, reqPerMin, conf.localConfig.ExpectedDuration)
		return nil, cmdx.FailSilently(cmd)
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

	memUsed, err := stats.LoadRawData(memStats).Max()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Unexpected maths error: %+v\nRaw Data: %+v\n", err, memStats)
	}

	return &resultTable{
		TotalTime:  totalTime,
		MedianTime: duration(calcData.Mean),
		StdDev:     duration(calcData.StandardDeviation),
		MinTime:    duration(calcData.Min),
		MaxTime:    duration(calcData.Max),
		MaxMem:     bytesize.ByteSize(memUsed),
	}, nil
}
