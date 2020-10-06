package schema

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"
)

type schema struct {
	name string
	raw  string
	s    *jsonschema.Schema
}

type schemas []schema

type result int

const (
	success result = iota
	failure
)

func (r result) String() string {
	return []string{"success", "failure"}[r]
}

func (s schema) validate(path string) error {
	if s.s == nil {
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(s.name, strings.NewReader(s.raw)); err != nil {
			return errors.WithStack(err)
		}

		sx, err := compiler.Compile(s.name)
		if err != nil {
			return errors.WithStack(err)
		}

		s.s = sx
	}

	var doc io.Reader
	y, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.WithStack(err)
	}

	j, err := yaml.YAMLToJSON(y)
	if err != nil {
		return errors.WithStack(err)
	}

	doc = bytes.NewBuffer(j)

	if err := s.s.Validate(doc); err != nil {
		return errors.Errorf("there were validation errors: %s", err)
	}

	return nil
}

func (ss *schemas) getByName(n string) (*schema, error) {
	for _, s := range *ss {
		if s.name == n {
			return &s, nil
		}
	}

	return nil, errors.Errorf("could not find schema with name %s", n)
}

func TestSchemas(t *testing.T) {
	t.Run("test .schema/config.schema.json", SchemaTestRunner("../../.schema", "config"))
}

func SchemaTestRunner(spath string, sname string) func(*testing.T) {
	return func(t *testing.T) {
		sb, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.schema.json", spath, sname))
		require.NoError(t, err)

		// To test refs independently and reduce test case size we replace every "$ref" with "const".
		// That way refs will not be resolved but we still make sure that they are pointing to the right definition.
		// Changing a definition will result in just changing test cases for that definition.
		s := strings.Replace(string(sb), `"$ref":`, `"const":`, -1)

		schemas := schemas{{
			name: "root",
			raw:  s,
		}}
		def := gjson.Get(s, "definitions")
		if def.Exists() {
			require.True(t, def.IsObject())
			def.ForEach(func(key, value gjson.Result) bool {
				require.Equal(t, gjson.String, key.Type)
				schemas = append(schemas, schema{
					name: key.String(),
					raw:  value.Raw,
				})
				return true
			})
		}

		RunCases(t, schemas, fmt.Sprintf("./fixtures/%s.schema.test.success", sname), success)
		RunCases(t, schemas, fmt.Sprintf("./fixtures/%s.schema.test.failure", sname), failure)
	}
}

func RunCases(t *testing.T, ss schemas, dir string, expected result) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if info.IsDir() {
			return nil
		}

		parts := strings.Split(info.Name(), ".")
		require.Equal(t, 3, len(parts))
		sName, tc := parts[0], parts[1]

		s, err := ss.getByName(sName)
		require.NoError(t, err)

		t.Run(fmt.Sprintf("case=schema %s test case %s expects %s", sName, tc, expected), func(t *testing.T) {
			err := s.validate(path)
			if expected == success {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})

		return nil
	})
	require.NoError(t, err)
}
