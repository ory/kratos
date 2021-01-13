package argon2

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/inhies/go-bytesize"
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
	c.c.DedicatedMemory = config.Argon2DefaultDedicatedMemory
	flags.Var(&c.c.DedicatedMemory, FlagDedicatedMemory, "Amount of memory dedicated for password hashing. Kratos will try to not consume more memory.")
	flags.DurationVar(&c.c.MinimalDuration, FlagMinimalDuration, config.Argon2DefaultDuration, "Minimal duration a hashing operation (~login request) takes.")
	flags.DurationVar(&c.c.ExpectedDeviation, FlagExpectedDeviation, config.Argon2DefaultDeviation, "Expected deviation of the time a hashing operation (~login request) takes.")

	flags.Uint8Var(&c.c.Parallelism, FlagParallelism, config.Argon2DefaultParallelism, "Number of threads to use.")

	flags.Uint32Var(&c.c.SaltLength, FlagSaltLength, config.Argon2DefaultSaltLength, "Length of the salt in bytes.")
	flags.Uint32Var(&c.c.KeyLength, FlagKeyLength, config.Argon2DefaultKeyLength, "Length of the key in bytes.")
}

func registerArgon2ConfigFlags(flags *pflag.FlagSet, c *argon2Config) {
	flags.Uint32Var(&c.c.Iterations, FlagIterations, 1, "Number of iterations to start probing at.")

	// set default value first
	c.memory = bytesize.ByteSize(config.Argon2DefaultMemory) * bytesize.KB
	flags.Var(&c.memory, FlagMemory, "Memory to use.")

	registerArgon2ConstantConfigFlags(flags, c)
}

func configProvider(cmd *cobra.Command, flagConf *argon2Config) (*argon2Config, error) {
	l := logrusx.New("ORY Kratos", config.Version)
	conf := &argon2Config{}
	var err error
	conf.config, err = config.New(l,
		configx.WithFlags(cmd.Flags()),
		configx.SkipValidation(),
		configx.WithContext(cmd.Context()),
		configx.WithImmutables("hashers"),
	)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Unable to initialize the config provider: %s\n", err.Error())
		return nil, cmdx.FailSilently(cmd)
	}
	c, err := conf.config.HasherArgon2()
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Unable to get the config from the provider: %s\n", err.Error())
		return nil, cmdx.FailSilently(cmd)
	}
	conf.c = *c

	if cmd.Flags().Changed(FlagIterations) {
		conf.c.Iterations = flagConf.c.Iterations
	}
	if cmd.Flags().Changed(FlagParallelism) {
		conf.c.Parallelism = flagConf.c.Parallelism
	}
	if cmd.Flags().Changed(FlagMemory) {
		conf.memory = flagConf.memory
	}
	if cmd.Flags().Changed(FlagDedicatedMemory) {
		conf.c.DedicatedMemory = flagConf.c.DedicatedMemory
	}
	if cmd.Flags().Changed(FlagKeyLength) {
		conf.c.KeyLength = flagConf.c.KeyLength
	}
	if cmd.Flags().Changed(FlagSaltLength) {
		conf.c.SaltLength = flagConf.c.SaltLength
	}
	if cmd.Flags().Changed(FlagExpectedDeviation) {
		conf.c.ExpectedDeviation = flagConf.c.ExpectedDeviation
	}
	if cmd.Flags().Changed(FlagMinimalDuration) {
		conf.c.MinimalDuration = flagConf.c.MinimalDuration
	}

	return conf, nil
}

type (
	argon2Config struct {
		c      config.HasherArgon2Config
		memory bytesize.ByteSize
		config *config.Provider
	}
)

var _ cmdx.TableRow = &argon2Config{}

func (c *argon2Config) Header() []string {
	var header []string

	t := reflect.TypeOf(c.c)
	for i := 0; i < t.NumField(); i++ {
		header = append(header, strings.ReplaceAll(strings.ToUpper(t.Field(i).Tag.Get("json")), "_", " "))
	}

	return header
}

func (c *argon2Config) Columns() []string {
	conf, _ := c.HasherArgon2()
	return []string{
		fmt.Sprintf("%d", conf.Memory),
		fmt.Sprintf("%d", conf.Iterations),
		fmt.Sprintf("%d", conf.Parallelism),
		fmt.Sprintf("%d", conf.SaltLength),
		fmt.Sprintf("%d", conf.KeyLength),
		conf.MinimalDuration.String(),
		conf.ExpectedDeviation.String(),
		conf.DedicatedMemory.String(),
	}
}

func (c *argon2Config) Interface() interface{} {
	i, _ := c.HasherArgon2()
	return i
}

func (c *argon2Config) Configuration(ctx context.Context) *config.Provider {
	ac, _ := c.HasherArgon2()
	for k, v := range map[string]interface{}{
		config.ViperKeyHasherArgon2ConfigIterations:        ac.Iterations,
		config.ViperKeyHasherArgon2ConfigMemory:            ac.Memory,
		config.ViperKeyHasherArgon2ConfigParallelism:       ac.Parallelism,
		config.ViperKeyHasherArgon2ConfigDedicatedMemory:   ac.DedicatedMemory,
		config.ViperKeyHasherArgon2ConfigKeyLength:         ac.KeyLength,
		config.ViperKeyHasherArgon2ConfigSaltLength:        ac.SaltLength,
		config.ViperKeyHasherArgon2ConfigMinimalDuration:   ac.MinimalDuration,
		config.ViperKeyHasherArgon2ConfigExpectedDeviation: ac.ExpectedDeviation,
	} {
		_ = c.config.Set(k, v)
	}
	return c.config
}

func (c *argon2Config) HasherArgon2() (*config.HasherArgon2Config, error) {
	if c.memory != 0 {
		c.c.Memory = uint32(c.memory / bytesize.KB)
	}
	if c.c.Memory == 0 {
		c.c.Memory = config.Argon2DefaultMemory
	}
	return &c.c, nil
}

func (c *argon2Config) getMemFormat() string {
	return (bytesize.ByteSize(c.c.Memory) * bytesize.KB).String()
}
