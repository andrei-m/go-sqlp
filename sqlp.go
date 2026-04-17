package sqlp

import (
	"fmt"
	"text/template"
)

// ExecutionConfig configures options for an Execution.
type ExecutionConfig struct {
	NumberedParameters bool
	FuncName           string
}

var DefaultExecutionConfig = ExecutionConfig{
	NumberedParameters: false,
	FuncName: "param",
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
	e.args = append(e.args, arg)
	if e.conf.NumberedParameters {
		return fmt.Sprintf("$%d", len(e.args))
	}
	return "?"
}
