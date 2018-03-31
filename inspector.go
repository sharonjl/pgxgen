// Copyright © 2018 Sharon Lourduraj
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pgxgen

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/reiver/go-stringcase"
)

var pgToPgxTypeMap = map[string]string{
	"text":        "pgtype.Text",
	"varchar":     "pgtype.Text",
	"bytea":       "pgtype.Bytea",
	"int2":        "pgtype.Int2",
	"int4":        "pgtype.Int4",
	"int8":        "pgtype.Int8",
	"bool":        "pgtype.Bool",
	"uuid":        "pgtype.UUID",
	"_uuid":       "pgtype.UUIDArray",
	"timestamp":   "pgtype.Timestamp",
	"timestamptz": "pgtype.Timestamptz",
	"float4":      "pgtype.Float4",
	"float8":      "pgtype.Float8",
	"jsonb":       "pgtype.JSONB",
}

var pgToGoTypeMap = map[string]string{
	"text":        "string",
	"varchar":     "string",
	"bytea":       "[]byte",
	"int2":        "int16",
	"int4":        "int32",
	"int8":        "int64",
	"bool":        "bool",
	"uuid":        "uuid.UUID",
	"_uuid":       "[]uuid.UUID",
	"timestamp":   "time.Time",
	"timestamptz": "time.Time",
	"float4":      "float32",
	"float8":      "float64",
	"jsonb":       "[]byte",
}

var pgToGoTemplate = map[string]func(v, p string) string{
	"text":        func(v, p string) string { return fmt.Sprintf("%s.%s.String", v, p) },
	"varchar":     func(v, p string) string { return fmt.Sprintf("%s.%s.String", v, p) },
	"bytea":       func(v, p string) string { return fmt.Sprintf("%s.%s.Bytes", v, p) },
	"int2":        func(v, p string) string { return fmt.Sprintf("%s.%s.Int", v, p) },
	"int4":        func(v, p string) string { return fmt.Sprintf("%s.%s.Int", v, p) },
	"int8":        func(v, p string) string { return fmt.Sprintf("%s.%s.Int", v, p) },
	"bool":        func(v, p string) string { return fmt.Sprintf("%s.%s.Bool", v, p) },
	"uuid":        func(v, p string) string { return fmt.Sprintf("uuid.FromBytesOrNil(%s.%s.Bytes[:])", v, p) },
	"_uuid":       func(v, p string) string { return fmt.Sprintf("ToUUIDSlice(%s.%s)", v, p) },
	"timestamp":   func(v, p string) string { return fmt.Sprintf("%s.%s.Time", v, p) },
	"timestamptz": func(v, p string) string { return fmt.Sprintf("%s.%s.Time", v, p) },
	"float4":      func(v, p string) string { return fmt.Sprintf("%s.%s.Float", v, p) },
	"float8":      func(v, p string) string { return fmt.Sprintf("%s.%s.Float", v, p) },
	"jsonb":       func(v, p string) string { return fmt.Sprintf("%s.%s.Bytes", v, p) },
}

var goToPgTemplate = map[string]func(v string) string{
	"text":        func(v string) string { return fmt.Sprintf("pgtype.Text{String: %s, Status: pgtype.Present}", v) },
	"varchar":     func(v string) string { return fmt.Sprintf("pgtype.Text{String: %s, Status: pgtype.Present}", v) },
	"bytea":       func(v string) string { return fmt.Sprintf("pgtype.Bytea{Bytes: %s, Status: pgtype.Present}", v) },
	"int2":        func(v string) string { return fmt.Sprintf("pgtype.Int2{Int: %s, Status: pgtype.Present}", v) },
	"int4":        func(v string) string { return fmt.Sprintf("pgtype.Int4{Int: %s, Status: pgtype.Present}", v) },
	"int8":        func(v string) string { return fmt.Sprintf("pgtype.Int8{Int: %s, Status: pgtype.Present}", v) },
	"bool":        func(v string) string { return fmt.Sprintf("pgtype.Bool{Bool: %s, Status: pgtype.Present}", v) },
	"uuid":        func(v string) string { return fmt.Sprintf("pgtype.UUID{Bytes: [16]byte(%s), Status: pgtype.Present}", v) },
	"_uuid":       func(v string) string { return fmt.Sprintf("UUIDArray(%s)", v) },
	"timestamp":   func(v string) string { return fmt.Sprintf("pgtype.Timestamp{Time: %s, Status: pgtype.Present}", v) },
	"timestamptz": func(v string) string { return fmt.Sprintf("pgtype.Timestamp{Time: %s, Status: pgtype.Present}", v) },
	"float4":      func(v string) string { return fmt.Sprintf("pgtype.Float4{Float: %s, Status: pgtype.Present}", v) },
	"float8":      func(v string) string { return fmt.Sprintf("pgtype.Float8{Float: %s, Status: pgtype.Present}", v) },
	"jsonb":       func(v string) string { return fmt.Sprintf("pgtype.JSONB{Bytes: %s, Status: pgtype.Present}", v) },
}

var recognizedAcronyms = map[string]string{
	"Id":  "ID",
	"Ip":  "IP",
	"Url": "URL",
	"Fb" : "FB",
}

var customEnumType = []string{}

func replaceAcronyms(s string) string {
	for k, v := range recognizedAcronyms {
		s = strings.Replace(s, k, v, 1)
	}
	return s
}

func ExportedName(s string) string {
	return replaceAcronyms(stringcase.ToPascalCase(s))
}

func shortName(s string) string {
	var r []rune
	for _, c := range s {
		if unicode.IsUpper(c) {
			r = append(r, unicode.ToLower(c))
		}
	}
	return string(r)
}

type Enum struct {
	Name   string
	Values []*EnumValue
}

func (e *Enum) ExportedName() string {
	return ExportedName(e.Name)
}

func (e *Enum) ShortName() string {
	return shortName(replaceAcronyms(stringcase.ToPascalCase(e.Name)))
}

func (e *Enum) PgxType() string {
	return pgToPgxTypeMap[e.Name]
}

func (e *Enum) GoType() string {
	return pgToGoTypeMap[e.Name]
}

type EnumValue struct {
	Value string
}

func (e *EnumValue) ExportedName() string {
	return ExportedName(e.Value)
}

func (e *EnumValue) GoType() string {
	return pgToGoTypeMap[e.Value]
}

type Table struct {
	Catalog     string
	Schema      string
	Name        string
	Columns     []*Column
	PrimaryKeys []*Column
	Indexes     map[string][]*Column
}

func (t *Table) ExportedName() string {
	return ExportedName(t.Name)
}

func (t *Table) ShortName() string {
	return shortName(replaceAcronyms(stringcase.ToPascalCase(t.Name)))
}

func (t *Table) GoType() string {
	return t.ExportedName()
}

type Column struct {
	Position int
	Nullable bool
	Name     string
	DataType string
	IsPK     bool
}

func (c *Column) ExportedName() string {
	return ExportedName(c.Name)
}

func (c *Column) ShortName() string {
	return shortName(replaceAcronyms(stringcase.ToPascalCase(c.Name)))
}

func (c *Column) PgxType() string {
	return pgToPgxTypeMap[c.DataType]
}

func (c *Column) GoType() string {
	return pgToGoTypeMap[c.DataType]
}

func (c *Column) GoVar() string {
	return replaceAcronyms(stringcase.ToCamelCase(c.Name))
}

func (c *Column) GoVarTemplate() string {
	for _, t := range customEnumType {
		if c.DataType == t {
			return "string(" + replaceAcronyms(stringcase.ToCamelCase(c.Name)) + ")"
		}
	}
	return replaceAcronyms(stringcase.ToCamelCase(c.Name))
}

func (c *Column) GoValueTemplate(v string) string {
	return pgToGoTemplate[c.DataType](v, c.ExportedName())
}

func (c *Column) PgValueTemplate(v string) string {
	return goToPgTemplate[c.DataType](v)
}

type PGData struct {
	Enums  map[string]*Enum
	Tables map[string]*Table
}

func Inspect(conn *pgx.Conn, schema string) (*PGData, error) {
	data := &PGData{}
	enums, err := getEnums(conn)
	if err != nil {
		return nil, errors.WithMessage(err, "querying enums")
	}
	data.Enums = enums
	for name, en := range enums {
		pgToPgxTypeMap[name] = "PGType" + en.ExportedName()
		pgToGoTypeMap[name] = en.ExportedName()
		pgToGoTemplate[name] = func(t string) func(v, p string) string {
			return func(v, p string) string { return fmt.Sprintf("%s(%s.%s.String)", t, v, p) }
		}(en.GoType())
		goToPgTemplate[name] = func(v string) string { return fmt.Sprintf("%s.PGType()", v) }
		customEnumType = append(customEnumType, name)
	}

	tables, err := getTables(conn, schema)
	if err != nil {
		return nil, errors.WithMessage(err, "querying tables")
	}
	data.Tables = tables
	for _, t := range tables {
		cols, err := getColumns(conn, schema, t.Name)
		if err != nil {
			return nil, errors.WithMessage(err, "")
		}
		t.Columns = cols

		pkColName, err := getTablePrimaryIndex(conn, schema, t.Name)
		if err != nil {
			return nil, errors.WithMessage(err, "")
		}
		t.PrimaryKeys = []*Column{}
		for _, c := range t.Columns {
			for _, pkCol := range pkColName {
				if c.Name == pkCol {
					c.IsPK = true
					t.PrimaryKeys = append(t.PrimaryKeys, c)
				}
			}
		}
	}
	return data, nil
}

const (
	queryGetTables = `
SELECT
  table_catalog,
  table_schema,
  table_name
FROM information_schema.tables
WHERE table_schema = $1;`

	queryGetTablePrimaryIndex = `
SELECT
  pg_attribute.attname
FROM pg_index, pg_class, pg_attribute, pg_namespace
WHERE
  pg_class.oid = $2::regclass AND
  indrelid = pg_class.oid AND
  nspname = $1 AND
  pg_class.relnamespace = pg_namespace.oid AND
  pg_attribute.attrelid = pg_class.oid AND
  pg_attribute.attnum = any(pg_index.indkey)
 AND indisprimary
`

	queryGetColumns = `
SELECT ordinal_position, column_name, udt_name, is_nullable
FROM information_schema.columns
WHERE table_schema = $1
  AND table_name   = $2;
`

	queryGetEnums = `
SELECT
  n.nspname   AS enum_schema,
  t.typname   AS enum_name,
  e.enumlabel AS enum_value
FROM pg_type t
  JOIN pg_enum e ON t.oid = e.enumtypid
  JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace;
`
	queryGetTableIndexes = `
SELECT
  i.relname AS index_name,
  idx.indrelid :: REGCLASS as table_name,
  ARRAY(
      SELECT pg_get_indexdef(idx.indexrelid, k + 1, TRUE)
      FROM generate_subscripts(idx.indkey, 1) AS k
      ORDER BY k
  )         AS index_columns
FROM pg_index AS idx
  JOIN pg_class AS i
    ON i.oid = idx.indexrelid
  JOIN pg_am AS am
    ON i.relam = am.oid
  JOIN pg_namespace AS ns
    ON ns.oid = i.relnamespace
       AND ns.nspname = ANY (current_schemas(FALSE));
`
)

func getEnums(conn *pgx.Conn) (map[string]*Enum, error) {
	rows, err := conn.Query(queryGetEnums)
	defer rows.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	enMap := map[string]*Enum{}
	for rows.Next() {
		var sch, name, value string
		err := rows.Scan(&sch, &name, &value)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	getEnum:
		en, ok := enMap[name]
		if !ok {
			enMap[name] = &Enum{
				Name:   name,
				Values: []*EnumValue{},
			}
			goto getEnum
		}
		en.Values = append(en.Values, &EnumValue{Value: value})
	}
	return enMap, nil
}

func getTables(conn *pgx.Conn, schema string) (map[string]*Table, error) {
	rows, err := conn.Query(queryGetTables, schema)
	defer rows.Close()
	if err != nil {
		return nil, errors.Errorf("unable to get tables: %v", err)
	}

	tables := make(map[string]*Table)
	for rows.Next() {
		var table Table
		err := rows.Scan(&table.Catalog, &table.Schema, &table.Name)
		if err != nil {
			return nil, err
		}
		tables[table.Name] = &table
	}
	return tables, nil
}

func getColumns(conn *pgx.Conn, schema, table string) ([]*Column, error) {
	rows, err := conn.Query(queryGetColumns, schema, table)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get columns: %v", err)
	}

	var cols []*Column
	for rows.Next() {
		var col Column
		var null string
		err := rows.Scan(&col.Position, &col.Name, &col.DataType, &null)
		if null == "YES" {
			col.Nullable = true
		}
		if null == "NO" {
			col.Nullable = false
		}
		if err != nil {
			return nil, err
		}
		cols = append(cols, &col)
	}
	return cols, nil
}

//
func getTablePrimaryIndex(conn *pgx.Conn, schema string, table string) ([]string, error) {
	rows, err := conn.Query(queryGetTablePrimaryIndex, schema, table)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get tables: %v", err)
	}

	var idx []string
	for rows.Next() {
		var colName string
		err := rows.Scan(&colName)
		if err != nil {
			return nil, err
		}
		idx = append(idx, colName)
	}
	return idx, nil
}

//
// func getTableIndexes(conn *pgx.Conn) ([]*Index, error) {
// 	rows, err := conn.Query(queryGetTableIndexes)
// 	defer rows.Close()
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to get tables: %v", err)
// 	}
//
// 	var idx []*Index
// 	for rows.Next() {
// 		var indexName string
// 		var tableName string
// 		var _indexCols pgtype.TextArray
// 		var indexCols []string
// 		err := rows.Scan(&indexName, &tableName, &_indexCols)
// 		if err != nil {
// 			return nil, err
// 		}
// 		_indexCols.AssignTo(&indexCols)
// 		idx = append(idx, &Index{
// 			Name:      indexName,
// 			GoName:    replaceAbbr(stringcase.ToPascalCase(indexName)),
// 			ShortName: shortName(replaceAbbr(stringcase.ToPascalCase(indexName))),
// 			Table:     tableName,
// 			Columns:   indexCols,
// 		})
// 	}
// 	return idx, nil
// }
