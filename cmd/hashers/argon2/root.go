package argon2

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
)

const (
	FlagIterations        = "iterations"
	FlagParallelism       = "parallelism"
	FlagSaltLength        = "salt-length"
	FlagKeyLength         = "key-length"
	FlagMemory            = "memory"
	FlagDedicatedMemory   = "dedicated-memory"
	FlagMinimalDuration   = "min-duration"
	FlagExpectedDeviation = "expected-deviation"
)

var rootCmd = &cobra.Command{
	Use: "argon2",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(rootCmd)

	rootCmd.AddCommand(newCalibrateCmd(), newHashCmd(), newLoadTestCmd())
}

func registerArgon2ConstantConfigFlags(flags *pflag.FlagSet, c *argon2Config) {
	// set default value first
	c.localConfig.DedicatedMemory = config.Argon2DefaultDedicatedMemory
	flags.Var(&c.localConfig.DedicatedMemory, FlagDedicatedMemory, "Amount of memory dedicated for password hashing. Kratos will try to not consume more memory.")
	flags.DurationVar(&c.localConfig.ExpectedDuration, FlagMinimalDuration, config.Argon2DefaultDuration, "Minimal duration a hashing operation (~login request) takes.")
	flags.DurationVar(&c.localConfig.ExpectedDeviation, FlagExpectedDeviation, config.Argon2DefaultDeviation, "Expected deviation of the time a hashing operation (~login request) takes.")

	flags.Uint8Var(&c.localConfig.Parallelism, FlagParallelism, config.Argon2DefaultParallelism, "Number of threads to use.")

	flags.Uint32Var(&c.localConfig.SaltLength, FlagSaltLength, config.Argon2DefaultSaltLength, "Length of the salt in bytes.")
	flags.Uint32Var(&c.localConfig.KeyLength, FlagKeyLength, config.Argon2DefaultKeyLength, "Length of the key in bytes.")
}

func registerArgon2ConfigFlags(flags *pflag.FlagSet, c *argon2Config) {
	flags.Uint32Var(&c.localConfig.Iterations, FlagIterations, 1, "Number of iterations to start probing at.")

	// set default value first
	c.localConfig.Memory = config.Argon2DefaultMemory
	flags.Var(&c.localConfig.Memory, FlagMemory, "Memory to use.")

	registerArgon2ConstantConfigFlags(flags, c)
}

func configProvider(cmd *cobra.Command, flagConf *argon2Config) (*argon2Config, error) {
	l := logrusx.New("Ory Kratos", config.Version)
	conf := &argon2Config{}
	var err error
	conf.config, err = config.New(
		context.Background(),
		l,
		configx.WithFlags(cmd.Flags()),
		configx.SkipValidation(),
		configx.WithContext(cmd.Context()),
		configx.WithImmutables("hashers"),
	)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Unable to initialize the config provider: %s\n", err)
		return nil, cmdx.FailSilently(cmd)
	}
	conf.localConfig = *conf.config.HasherArgon2()

	if cmd.Flags().Changed(FlagIterations) {
		conf.localConfig.Iterations = flagConf.localConfig.Iterations
	}
	if cmd.Flags().Changed(FlagParallelism) {
		conf.localConfig.Parallelism = flagConf.localConfig.Parallelism
	}
	if cmd.Flags().Changed(FlagMemory) {
		conf.localConfig.Memory = flagConf.localConfig.Memory
	}
	if cmd.Flags().Changed(FlagDedicatedMemory) {
		conf.localConfig.DedicatedMemory = flagConf.localConfig.DedicatedMemory
	}
	if cmd.Flags().Changed(FlagKeyLength) {
		conf.localConfig.KeyLength = flagConf.localConfig.KeyLength
	}
	if cmd.Flags().Changed(FlagSaltLength) {
		conf.localConfig.SaltLength = flagConf.localConfig.SaltLength
	}
	if cmd.Flags().Changed(FlagExpectedDeviation) {
		conf.localConfig.ExpectedDeviation = flagConf.localConfig.ExpectedDeviation
	}
	if cmd.Flags().Changed(FlagMinimalDuration) {
		conf.localConfig.ExpectedDuration = flagConf.localConfig.ExpectedDuration
	}

	return conf, nil
}

type (
	argon2Config struct {
		localConfig config.Argon2
		config      *config.Config
	}
)

var _ cmdx.TableRow = &argon2Config{}

func (c *argon2Config) Header() []string {
	var header []string

	t := reflect.TypeOf(c.localConfig)
	for i := 0; i < t.NumField(); i++ {
		header = append(header, strings.ReplaceAll(strings.ToUpper(t.Field(i).Tag.Get("json")), "_", " "))
	}

	return header
}

func (c *argon2Config) Columns() []string {
	conf, _ := c.HasherArgon2()
	return []string{
		conf.Memory.String(),
		fmt.Sprintf("%d", conf.Iterations),
		fmt.Sprintf("%d", conf.Parallelism),
		fmt.Sprintf("%d", conf.SaltLength),
		fmt.Sprintf("%d", conf.KeyLength),
		conf.ExpectedDuration.String(),
		conf.ExpectedDeviation.String(),
		conf.DedicatedMemory.String(),
	}
}

func (c *argon2Config) Interface() interface{} {
	i, _ := c.HasherArgon2()
	return i
}

func (c *argon2Config) Config(_ context.Context) *config.Config {
	ac, _ := c.HasherArgon2()
	for k, v := range map[string]interface{}{
		config.ViperKeyHasherArgon2ConfigIterations:        ac.Iterations,
		config.ViperKeyHasherArgon2ConfigMemory:            ac.Memory,
		config.ViperKeyHasherArgon2ConfigParallelism:       ac.Parallelism,
		config.ViperKeyHasherArgon2ConfigDedicatedMemory:   ac.DedicatedMemory,
		config.ViperKeyHasherArgon2ConfigKeyLength:         ac.KeyLength,
		config.ViperKeyHasherArgon2ConfigSaltLength:        ac.SaltLength,
		config.ViperKeyHasherArgon2ConfigExpectedDuration:  ac.ExpectedDuration,
		config.ViperKeyHasherArgon2ConfigExpectedDeviation: ac.ExpectedDeviation,
	} {
		_ = c.config.Set(k, v)
	}
	return c.config
}

func (c *argon2Config) HasherArgon2() (*config.Argon2, error) {
	if c.localConfig.Memory == 0 {
		c.localConfig.Memory = config.Argon2DefaultMemory
	}
	return &c.localConfig, nil
}
