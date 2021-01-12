package argon2

import (
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
	FlagIterations      = "iterations"
	FlagParallelism     = "parallelism"
	FlagSaltLength      = "salt-length"
	FlagKeyLength       = "key-length"
	FlagMemory          = "memory"
	FlagDedicatedMemory = "dedicated-memory"
)

var rootCmd = &cobra.Command{
	Use: "argon2",
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(rootCmd)

	rootCmd.AddCommand(newCalibrateCmd(), newHashCmd(), newLoadTestCmd())
}

func registerArgon2ConstantConfigFlags(flags *pflag.FlagSet, c *argon2Config) {
	flags.Uint8Var(&c.c.Parallelism, FlagParallelism, config.Argon2DefaultParallelism, "Number of threads to use.")
	flags.Var(&c.c.DedicatedMemory, FlagDedicatedMemory, "Amount of memory dedicated for password hashing. Kratos will try to not consume more memory.")

	flags.Uint32Var(&c.c.SaltLength, FlagSaltLength, config.Argon2DefaultSaltLength, "Length of the salt in bytes.")
	flags.Uint32Var(&c.c.KeyLength, FlagKeyLength, config.Argon2DefaultKeyLength, "Length of the key in bytes.")
}

func registerArgon2ConfigFlags(flags *pflag.FlagSet, c *argon2Config) {
	flags.Uint32Var(&c.c.Iterations, FlagIterations, 1, "Number of iterations to start probing at.")
	flags.Var(&c.memory, FlagMemory, "Memory to use.")

	registerArgon2ConstantConfigFlags(flags, c)
}

func configProvider(cmd *cobra.Command, flagConf *argon2Config) (*argon2Config, error) {
	l := logrusx.New("ORY Kratos", config.Version)
	c, err := config.New(l,
		configx.WithFlags(cmd.Flags()),
		configx.SkipValidation(),
		configx.WithContext(cmd.Context()),
		configx.WithImmutables("hashers"),
	)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Unable to initialize the config provider: %s\n", err.Error())
		return nil, cmdx.FailSilently(cmd)
	}

	cv, _ := c.HasherArgon2()
	conf := &argon2Config{c: *cv}
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

	return conf, nil
}

type (
	argon2Config struct {
		c      config.HasherArgon2Config
		memory bytesize.ByteSize
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
		fmt.Sprintf("%s", conf.MinimalDuration),
		fmt.Sprintf("%s", conf.ExpectedDeviation),
		fmt.Sprintf("%s", conf.DedicatedMemory),
	}
}

func (c *argon2Config) Interface() interface{} {
	i, _ := c.HasherArgon2()
	return i
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
