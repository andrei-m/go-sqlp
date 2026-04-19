package sqlp

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

// ExecutionConfig configures options for an Execution.
type ExecutionConfig struct {
	NumberedParameters bool
	FuncName           string
}

var DefaultExecutionConfig = ExecutionConfig{
	NumberedParameters: false,
	FuncName:           "param",
}

// NewExecution creates an Execution for use with a single text/template
// execution using provided configuration options.
func NewExecution(conf ExecutionConfig) *Execution {
	return &Execution{
		conf: conf,
		args: []any{},
	}
}

// Execution provides the "param" function for execution in templates
// and collects positional parameter values.
//
// The "param" function handles different types as follows:
//
//   - Slices (e.g., []int, []string): The slice is "unrolled" into a comma-separated
//     list of placeholders (e.g., "?, ?, ?"), and each element is added as a separate
//     argument. This is useful for "IN" clauses.
//   - []byte: Treated as a single atomic parameter, not unrolled.
//   - All other types: Treated as a single atomic parameter.
type Execution struct {
	conf ExecutionConfig
	args []any
}

// Args returns query arguments for use with database/sql query functions.
func (e *Execution) Args() []any {
	return e.args
}

// Func provides the "param" function to be included in a template.FuncMap for use with Template.Funcs
func (e *Execution) Func() (string, any) {
	return e.conf.FuncName, e.param
}

// Funcs provides a map containing the "param" function for use with Template.Funcs
func (e *Execution) Funcs() template.FuncMap {
	return template.FuncMap{
		e.conf.FuncName: e.param,
	}
}

func (e *Execution) param(arg any) string {
	v := reflect.ValueOf(arg)
	// []byte (slice of uint8) should be treated as a single parameter.
	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8 {
		var positionalParams []string
		for i := 0; i < v.Len(); i++ {
			e.args = append(e.args, v.Index(i).Interface())
			positionalParams = append(positionalParams, e.nextPositionalParameter())
		}
		return strings.Join(positionalParams, ", ")
	}

	e.args = append(e.args, arg)
	return e.nextPositionalParameter()
}

func (e *Execution) nextPositionalParameter() string {
	if e.conf.NumberedParameters {
		return fmt.Sprintf("$%d", len(e.args))
	}
	return "?"
}
