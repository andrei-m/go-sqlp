package sqlp

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PositionalParameters(t *testing.T) {
	type Params struct {
		Foo string
		Bar int
	}

	sqlTemplate := `
SELECT *
FROM table
WHERE true
{{if .Foo}}
AND foo = {{param .Foo}}
{{- end -}}
{{if .Bar}}
AND bar = {{param .Bar}}
{{- end -}}`

	t.Run("'foo' not empty", func(t *testing.T) {
		p := Params{Foo: "value"}
		exec := NewExecution(DefaultExecutionConfig)
		sql, err := executeTemplate(sqlTemplate, p, exec.Funcs())
		require.NoError(t, err)
		assert.Equal(t, `
SELECT *
FROM table
WHERE true

AND foo = ?`, sql)
		assert.Equal(t, []any{"value"}, exec.Args())
	})

	t.Run("'bar' not empty", func(t *testing.T) {
		p := Params{Bar: 42}
		exec := NewExecution(DefaultExecutionConfig)
		sql, err := executeTemplate(sqlTemplate, p, exec.Funcs())
		require.NoError(t, err)
		assert.Equal(t, `
SELECT *
FROM table
WHERE true

AND bar = ?`, sql)
		assert.Equal(t, []any{42}, exec.Args())
	})

	t.Run("'foo' and 'bar' not empty", func(t *testing.T) {
		p := Params{Foo: "value", Bar: 42}
		exec := NewExecution(DefaultExecutionConfig)
		sql, err := executeTemplate(sqlTemplate, p, exec.Funcs())
		require.NoError(t, err)
		assert.Equal(t, `
SELECT *
FROM table
WHERE true

AND foo = ?
AND bar = ?`, sql)
		assert.Equal(t, []any{"value", 42}, exec.Args())
	})

	t.Run("numbered parameters", func(t *testing.T) {
		p := Params{Foo: "value", Bar: 42}
		exec := NewExecution(ExecutionConfig{
			NumberedParameters: true,
			FuncName:           "param",
		})
		sql, err := executeTemplate(sqlTemplate, p, exec.Funcs())
		require.NoError(t, err)
		assert.Equal(t, `
SELECT *
FROM table
WHERE true

AND foo = $1
AND bar = $2`, sql)
		assert.Equal(t, []any{"value", 42}, exec.Args())
	})
}

func executeTemplate(sqlTemplate string, params any, funcs template.FuncMap) (string, error) {
	tmpl, err := template.New("sql template").Funcs(funcs).Parse(sqlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	b := bytes.Buffer{}
	if err := tmpl.Execute(&b, params); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return b.String(), nil
}
