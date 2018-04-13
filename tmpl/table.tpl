// Code generated by pgxgen. DO NOT EDIT.
package {{.PackageName}}

import (
    "time"
    "error"
	"strconv"
	"strings"
	"time"

    pgx "github.com/jackc/pgx"
    pgtype "github.com/jackc/pgx/pgtype"
    uuid "github.com/satori/go.uuid"
)

// {{.Table.ExportedName}} represents row data from the table '{{.Table.Name}}.'
type {{.Table.ExportedName}} struct {
{{range .Table.Columns -}}
    {{.ExportedName}} {{.PgxType}} // column: '{{.Name}}'
{{end -}}
}

{{range .Table.Columns}}
// Get{{.ExportedName}} returns value of column '{{.Name}}' as {{.GoType}}
func({{$.Table.ShortName}} *{{$.Table.GoType}}) Get{{.ExportedName}}() {{.GoType}} { return {{.GoValueTemplate $.Table.ShortName}} }

// Set{{.ExportedName}} sets {{.GoType}} value for column '{{.Name}}'
func({{$.Table.ShortName}} *{{$.Table.GoType}}) Set{{.ExportedName}}(v {{.GoType}}) *{{$.Table.GoType}} {
    {{$.Table.ShortName}}.{{.ExportedName}} = {{.PgValueTemplate "v"}}
    return {{$.Table.ShortName}}
}
{{end}}

//
const (
    Table{{.Table.ExportedName}} = "{{.Table.Name}}"
{{range .Table.Columns -}}
    Field{{$.Table.ExportedName}}{{.ExportedName}} = "{{.Name}}"
{{end -}}
)

// All{{.Table.ExportedName}}FieldsSlice is a slice of all field names for table '{{.Table.Name}}.'
var All{{.Table.ExportedName}}FieldsSlice = []string{
{{- range .Table.Columns}}
    "{{.Name}}",
{{- end}}
}

// All{{.Table.ExportedName}}FieldsStr is a comma separated string of all field names for table '{{.Table.Name}}.'
const All{{.Table.ExportedName}}FieldsStr = "{{range $k, $e := .Table.Columns}}{{if $k}}, {{end}}{{.Name}}{{end}}"

// Scan{{.Table.ExportedName}}s returns a single row containing all fields of '{{.Table.Name}}.' Reading of columns
// from result set is positional, and in the following order:
{{- range .Table.Columns}}
//      {{.Name}}
{{- end}}
func Scan{{.Table.ExportedName}}(row *pgx.Row) (*{{.Table.ExportedName}}, error) {
	m := &{{.Table.ExportedName}}{}
	err := row.Scan(
		{{- range .Table.Columns}}
            &m.{{.ExportedName}},
        {{- end}}
	)
	return m, err
}

// Scan{{.Table.ExportedName}}s returns one or more rows containing all fields of '{{.Table.Name}}.' Reading of columns
// from result set is positional, and in the following order:
{{- range .Table.Columns}}
//      {{.Name}}
{{- end}}
func Scan{{.Table.ExportedName}}s(rows *pgx.Rows) ([]*{{.Table.ExportedName}}, error) {
	var r []*{{.Table.ExportedName}}
	var err error
	for rows.Next() && err == nil {
		m := &{{.Table.ExportedName}}{}
		err = rows.Scan(
		{{- range .Table.Columns}}
            &m.{{.ExportedName}},
        {{- end}}
		)
		r = append(r, m)
	}
	return r, err
}

// Create{{.Table.ExportedName}} create a single row in '{{.Table.Name}}' and return it.
func Create{{.Table.ExportedName}}(db Conn, m *{{.Table.ExportedName}}) (*{{.Table.ExportedName}}, error) {
	var f []string
	var v []string
	var c int
	var a []interface{}

    {{range .Table.Columns}}
        if m.{{.ExportedName}}.Status != pgtype.Undefined {
            c++
            f = append(f, "{{.Name}}")
            v = append(v, "$"+strconv.Itoa(c))
            a = append(a, &m.{{.ExportedName}})
        }
    {{- end}}

	q := "INSERT INTO {{.Table.Schema}}.{{.Table.Name}} (" + strings.Join(f, ", ") + ") VALUES(" + strings.Join(v, ", ") + ") RETURNING " + All{{.Table.ExportedName}}FieldsStr + ";"

	row := db.QueryRow(q, a...)
	r, err := Scan{{.Table.ExportedName}}(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r, nil
}

// Update{{.Table.ExportedName}} updates a row in '{{.Table.Name}}.'
func Update{{.Table.ExportedName}}(db Conn, {{range $k, $pk := .Table.PrimaryKeys}}{{if $k}}, {{end}}{{.GoVar}} {{.GoType}}{{end}}, m *{{.Table.ExportedName}}) (*{{.Table.ExportedName}}, error) {
	var f []string
	var pk []string
	var c int
	var a []interface{}

    {{range .Table.PrimaryKeys}}
        { // Primary Key: {{.Name}}
        		c++
        		pk = append(pk, "{{.Name}} = $"+strconv.Itoa(c))
        		a = append(a, {{.GoVarTemplate}})
        }
    {{- end}}

    {{range .Table.Columns}}
        {{- if not .IsPK}}
        if m.{{.ExportedName}}.Status != pgtype.Undefined {
            c++
            f = append(f, "{{.Name}} = $"+strconv.Itoa(c))
            a = append(a, &m.{{.ExportedName}})
        }
        {{- end}}
    {{- end}}


	q := "UPDATE {{.Table.Schema}}.{{.Table.Name}} SET " + strings.Join(f, ", ") + " WHERE " + strings.Join(pk, " AND ") + " RETURNING " + All{{.Table.ExportedName}}FieldsStr + ";"
	row := db.QueryRow(q, a...)
	r, err := Scan{{.Table.ExportedName}}(row)
	return r, err
}

// Get{{.Table.ExportedName}} returns a row from '{{.Table.Name}}.' identified by primary key.
func Get{{.Table.ExportedName}}(db Conn, {{range $k, $pk := .Table.PrimaryKeys}}{{if $k}}, {{end}}{{.GoVar}} {{.GoType}}{{end}}) (*{{.Table.ExportedName}}, error) {
	q := "SELECT " + All{{.Table.ExportedName}}FieldsStr + " FROM {{.Table.Schema}}.{{.Table.Name}} WHERE {{range $k, $pk := .Table.PrimaryKeys}}{{if $k}} AND {{end}}{{.Name}} = ${{inc $k}}{{end}};"

	row := db.QueryRow(q, {{range $k, $pk := .Table.PrimaryKeys}}{{if $k}}, {{end}}{{.GoVarTemplate}}{{end}})
	r, err := Scan{{.Table.ExportedName}}(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r, nil
}

// Delete{{.Table.ExportedName}} returns a row from '{{.Table.Name}}.' identified by primary key.
func Delete{{.Table.ExportedName}}(db Conn, {{range $k, $pk := .Table.PrimaryKeys}}{{if $k}}, {{end}}{{.GoVar}} {{.GoType}}{{end}}) error {
	q := "DELETE FROM {{.Table.Schema}}.{{.Table.Name}} WHERE {{range $k, $pk := .Table.PrimaryKeys}}{{if $k}} AND {{end}}{{.Name}} = ${{inc $k}}{{end}};"
	_, err := db.Exec(q, {{range $k, $pk := .Table.PrimaryKeys}}{{if $k}}, {{end}}{{.GoVarTemplate}}{{end}})
    if err != nil {
        return err
    }
    return nil
}
