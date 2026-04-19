go-sqlp [![Godoc](https://godoc.org/github.com/andrei-m/go-sqlp?status.svg)](https://godoc.org/github.com/andrei-m/go-sqlp)
=======

**Development In Progress: sqlp's public API may change prior to version 1.0.0**

Go SQL Parameters works with Go's [text/template](https://pkg.go.dev/text/template) package to build parameterized SQL queries from templates and input context. Evaluating a template with go-sqlp produces parameterized SQL and a corresponding slice of input parameters ready to pass to `db.Query`.

Example with positional parameters:

```golang
type Params struct{
    Foo string
    Bar int
}

sqlTemplate := `
SELECT *
FROM table
WHERE true
{{if .Foo}}
AND foo = {{param .Foo}}
{{end}}
{{if .Bar}}
AND bar = {{param .Bar}}
{{end}}
`
p := Params{Foo: "value"}

exec := sqlp.NewExecution(sqlp.DefaultExecutionConfig)
// sqlp.DefaultFuncs binds `param` with defaults
tmpl, err := template.New("sql template").Funcs(exec.Funcs()).Parse(sqlTemplate)
if err != nil {
    return fmt.Errorf("failed to parse template: %w", err)
}

b := bytes.Buffer{}
if err := tmpl.Execute(&b, p); err != nil {
    return fmt.Errorf("failed to execute template: %w", err)
}

/*
SQL query evaluates to:
SELECT *
FROM table
WHERE true
AND foo = ?
*/
query := b.String()

// params evaluates to ["value"]
args := sqlp.Args()

row := db.QueryRow(query, args)
```

### Slice Unrolling

`param` automatically unrolls slices (except `[]byte`) into individual placeholders, making it easy to use with `IN` clauses:

```golang
type Params struct {
    IDs []int
}
p := Params{IDs: []int{1, 2, 3}}

// Template: "SELECT * FROM users WHERE id IN ({{param .IDs}})"
// Evaluates to: "SELECT * FROM users WHERE id IN (?, ?, ?)"
// Args: []any{1, 2, 3}
```

## Installation

```bash
go get github.com/andrei-m/go-sqlp
```

## Features

* `?` or ordinal numbered (`$1`, `$2`, etc.) SQL parameter syntax
* The `param` template function name can be changed
