package reltest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type Registry struct {
	names       []string
	valueToName map[interface{}]string
	nameToValue map[string]interface{}

	nameToYAML map[string]string
}

func NewRegistry() *Registry {
	__antithesis_instrumentation__.Notify(579063)
	return &Registry{
		valueToName: make(map[interface{}]string),
		nameToValue: make(map[string]interface{}),
		nameToYAML:  make(map[string]string),
	}
}

func (r *Registry) Register(name string, v interface{}) interface{} {
	__antithesis_instrumentation__.Notify(579064)
	if existing, exists := r.nameToValue[name]; exists {
		__antithesis_instrumentation__.Notify(579066)
		panic(errors.AssertionFailedf(
			"entity with name %s already registered %v, trying to register %v",
			name, existing, v,
		))
	} else {
		__antithesis_instrumentation__.Notify(579067)
	}
	__antithesis_instrumentation__.Notify(579065)
	r.names = append(r.names, name)
	r.nameToValue[name] = v
	r.valueToName[v] = name
	return v
}

func (r *Registry) FromYAML(name, yamlData string, dest interface{}) interface{} {
	__antithesis_instrumentation__.Notify(579068)
	if err := yaml.NewDecoder(strings.NewReader(yamlData)).Decode(dest); err != nil {
		__antithesis_instrumentation__.Notify(579070)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579071)
	}
	__antithesis_instrumentation__.Notify(579069)
	r.Register(name, dest)
	r.nameToYAML[name] = yamlData
	return dest
}

func (r *Registry) MustGetByName(t *testing.T, k string) interface{} {
	__antithesis_instrumentation__.Notify(579072)
	got, ok := r.nameToValue[k]
	require.Truef(t, ok, "MustGetByName(%s)", k)
	return got
}

func (r *Registry) MustGetName(t *testing.T, v interface{}) string {
	__antithesis_instrumentation__.Notify(579073)
	got, ok := r.GetName(v)
	require.Truef(t, ok, "MustGetName(%v)", v)
	return got
}

func (r *Registry) GetName(i interface{}) (string, bool) {
	__antithesis_instrumentation__.Notify(579074)
	got, ok := r.valueToName[i]
	return got, ok
}

func (r *Registry) valueToYAML(t *testing.T, name string) *yaml.Node {
	__antithesis_instrumentation__.Notify(579075)
	if yamlStr, hasToYAML := r.nameToYAML[name]; hasToYAML {
		__antithesis_instrumentation__.Notify(579077)
		var v interface{}
		require.NoError(t, yaml.NewDecoder(strings.NewReader(yamlStr)).Decode(&v))
		var n yaml.Node
		require.NoError(t, n.Encode(v))
		n.Style = yaml.FlowStyle
		return &n
	} else {
		__antithesis_instrumentation__.Notify(579078)
	}
	__antithesis_instrumentation__.Notify(579076)
	return r.EncodeToYAML(t, r.MustGetByName(t, name))
}

type RegistryYAMLEncoder interface {
	EncodeToYAML(t *testing.T, r *Registry) interface{}
}

func (r *Registry) EncodeToYAML(t *testing.T, v interface{}) *yaml.Node {
	__antithesis_instrumentation__.Notify(579079)
	toEncode := v
	if encoder, ok := v.(RegistryYAMLEncoder); ok {
		__antithesis_instrumentation__.Notify(579081)
		toEncode = encoder.EncodeToYAML(t, r)
	} else {
		__antithesis_instrumentation__.Notify(579082)
	}
	__antithesis_instrumentation__.Notify(579080)
	var n yaml.Node
	require.NoError(t, n.Encode(toEncode))
	return &n
}
