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
	"strings"

)

type QueryDefinitions struct {
	Query []QueryDefinition
}

type QueryDefinition struct {
	Name   string
	Table  string
	Fields []string
	Sort   []string
	Return string
}

type Query struct {
	Name       string
	Table      Table
	Filter     []Filter
	Sort       []Sort
	ReturnOne  bool
	ReturnMany bool
	Paged      bool
}

func (q *Query) ExportedName() string {
	return ExportedName(q.Name)
}

type Filter struct {
	Column Column
	Op     string
}

func (f Filter) SQLOp() string {
	switch f.Op {
	case "eq":
		return "="
	case "lt":
		return "<"
	case "lteq":
		return "<="
	case "gt":
		return ">"
	case "gteq":
		return ">="
	case "ne":
		return "!="
	}
	return "__OP__"
}

type Sort struct {
	Column Column
	Desc   bool
}

func (s Sort) String() string {
	if s.Desc {
		return "DESC"
	}
	return "ASC"
}

func ProcessQueryDefinitions(def QueryDefinitions, data PGData) []Query {
	var qq []Query
	for _, d := range def.Query {
		q := Query{Name: d.Name}
		q.Table = *data.Tables[d.Table]
		q.Filter = []Filter{}
		couldReturnMany := false
		for _, f := range d.Fields {
			ff := strings.Split(f, ":")
			if len(ff) != 2 {
				ff = []string{f, "eq"}
			}
			for _, c := range q.Table.Columns {
				if c.Name != ff[0] {
					continue
				}
				// Ops such as <, <=, >, >= could return many rows.
				couldReturnMany = couldReturnMany || !(ff[1] == "eq" || ff[1] == "ne")
				q.Filter = append(q.Filter, Filter{Column: *c, Op: ff[1]})
			}
		}

		q.Sort = []Sort{}
		for _, f := range d.Sort {
			for _, c := range q.Table.Columns {
				if c.Name == f || strings.HasSuffix(f, c.Name) {
					q.Sort = append(q.Sort, Sort{Column: *c, Desc: strings.HasPrefix(f, "-")})
				}
			}
		}
		q.ReturnOne = d.Return == "one"
		q.ReturnMany = couldReturnMany || d.Return == "many" || d.Return == "paged"
		q.Paged = d.Return == "paged"
		qq = append(qq, q)
	}

	// Add primary key for tables
	for _, t := range data.Tables {
		q := Query{Name: "Get" + t.ExportedName()}
		q.Table = *t
		q.Filter = []Filter{}
		for _, pk := range t.PrimaryKeys {
			q.Filter = append(q.Filter, Filter{
				Column: *pk,
				Op: "eq",
			})
		}
		q.Sort = []Sort{}
		q.ReturnOne = true
		q.ReturnMany = false
		q.Paged = false
		qq = append(qq, q)
	}
	return qq
}
