package sqlp

import (
	"bytes"
	"database/sql/driver"
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

func Test_SliceParameters(t *testing.T) {
	type Params struct {
		IDs  []int
		Tags []string
		Data []byte
		Role string
	}

	t.Run("unrolls int slice with positional parameters", func(t *testing.T) {
		p := Params{IDs: []int{1, 2, 3}}
		exec := NewExecution(DefaultExecutionConfig)
		sql, err := executeTemplate("WHERE id IN ({{param .IDs}})", p, exec.Funcs())
		require.NoError(t, err)
		assert.Equal(t, "WHERE id IN (?, ?, ?)", sql)
		assert.Equal(t, []any{1, 2, 3}, exec.Args())
	})

	t.Run("unrolls string slice with numbered parameters", func(t *testing.T) {
		p := Params{Tags: []string{"a", "b"}}
		exec := NewExecution(ExecutionConfig{NumberedParameters: true, FuncName: "param"})
		sql, err := executeTemplate("WHERE tag IN ({{param .Tags}})", p, exec.Funcs())
		require.NoError(t, err)
		assert.Equal(t, "WHERE tag IN ($1, $2)", sql)
		assert.Equal(t, []any{"a", "b"}, exec.Args())
	})

	t.Run("mixed atomic and slice with numbered parameters", func(t *testing.T) {
		p := Params{Role: "admin", IDs: []int{10, 20}}
		exec := NewExecution(ExecutionConfig{NumberedParameters: true, FuncName: "param"})
		sql, err := executeTemplate("WHERE role = {{param .Role}} AND id IN ({{param .IDs}})", p, exec.Funcs())
		require.NoError(t, err)
		assert.Equal(t, "WHERE role = $1 AND id IN ($2, $3)", sql)
		assert.Equal(t, []any{"admin", 10, 20}, exec.Args())
	})

	t.Run("does not unroll []byte", func(t *testing.T) {
		p := Params{Data: []byte("binary data")}
		exec := NewExecution(DefaultExecutionConfig)
		sql, err := executeTemplate("UPDATE t SET d = {{param .Data}}", p, exec.Funcs())
		require.NoError(t, err)
		assert.Equal(t, "UPDATE t SET d = ?", sql)
		assert.Equal(t, []any{[]byte("binary data")}, exec.Args())
	})

	t.Run("empty slice", func(t *testing.T) {
		p := Params{IDs: []int{}}
		exec := NewExecution(DefaultExecutionConfig)
		// Empty slice 'truthiness' should be used in a conditional block to avoid generating invalid SQL like "IN ()".
		sql, err := executeTemplate("IN ({{param .IDs}})", p, exec.Funcs())
		require.NoError(t, err)
		assert.Equal(t, "IN ()", sql)
		assert.Empty(t, exec.Args())
	})
}

type mockValuerSlice []string

func (m mockValuerSlice) Value() (driver.Value, error) {
	return "mocked", nil
}

func Test_Valuer(t *testing.T) {
	t.Run("does not unroll types implementing driver.Valuer", func(t *testing.T) {
		arg := mockValuerSlice{"a", "b"}
		exec := NewExecution(DefaultExecutionConfig)
		sql, err := executeTemplate("{{param .}}", arg, exec.Funcs())
		require.NoError(t, err)
		assert.Equal(t, "?", sql)
		assert.Equal(t, []any{arg}, exec.Args())
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
