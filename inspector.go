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
	"text":        "pgType.Text",
	"varchar":     "pgType.Text",
	"bytea":       "pgType.Bytea",
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
	"bytea":       func(v, p string) string { return fmt.Sprintf("%s.%s.String", v, p) },
	"int2":        func(v, p string) string { return fmt.Sprintf("%s.%s.Int", v, p) },
	"int4":        func(v, p string) string { return fmt.Sprintf("%s.%s.Int", v, p) },
	"int8":        func(v, p string) string { return fmt.Sprintf("%s.%s.Int", v, p) },
	"bool":        func(v, p string) string { return fmt.Sprintf("%s.%s.Bool", v, p) },
	"uuid":        func(v, p string) string { return fmt.Sprintf("uuid.FromBytesOrNil(%s.%s).Bytes[:]", v, p) },
	"_uuid":       func(v, p string) string { return fmt.Sprintf("ToUUIDSlice(%s.%s)", v, p) },
	"timestamp":   func(v, p string) string { return fmt.Sprintf("%s.%s.Time", v, p) },
	"timestamptz": func(v, p string) string { return fmt.Sprintf("%s.%s.Time", v, p) },
	"float4":      func(v, p string) string { return fmt.Sprintf("%s.%s.Float", v, p) },
	"float8":      func(v, p string) string { return fmt.Sprintf("%s.%s.Float", v, p) },
	"jsonb":       func(v, p string) string { return fmt.Sprintf("%s.%s.Bytes", v, p) },
}

var recognizedAcronyms = map[string]string{
	"id":  "ID",
	"ip":  "IP",
	"url": "URL",
}

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
	EnumName string
	Values   []string
}

func (e *Enum) ExportedName() string {
	return replaceAcronyms(stringcase.ToPascalCase(e.EnumName))
}

func (e *Enum) ShortName() string {
	return shortName(replaceAcronyms(stringcase.ToPascalCase(e.EnumName)))
}

func (e *Enum) PgxType() string {
	return pgToPgxTypeMap[e.EnumName]
}

func (e *Enum) GoType() string {
	return pgToGoTypeMap[e.EnumName]
}

type PGData struct {
	Enums map[string]*Enum
}

func Inspect(conn *pgx.Conn) (*PGData, error) {
	data := &PGData{}
	enums, err := getEnums(conn)
	if err != nil {
		return nil, errors.WithMessage(err, "querying enums")
	}
	data.Enums = enums
	for _, en := range enums {
		pgToPgxTypeMap[en.EnumName] = "PGType" + en.ExportedName()
		pgToGoTypeMap[en.EnumName] = en.ExportedName()
		pgToGoTemplate[en.EnumName] = func(v, p string) string { return fmt.Sprintf("%s(%s.%s).String", en.PgxType(), v, p) }
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
				EnumName: name,
				Values:   []string{},
			}
			goto getEnum
		}
		en.Values = append(en.Values, value)
	}
	return enMap, nil
}
