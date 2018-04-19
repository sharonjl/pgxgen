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

package pgconv

import (
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/satori/go.uuid"
)

func Text(s string) pgtype.Text { return pgtype.Text{String: s, Status: pgtype.Present} }
func Bool(b bool) pgtype.Bool   { return pgtype.Bool{Bool: b, Status: pgtype.Present} }

func Varchar(s string) pgtype.Varchar { return pgtype.Varchar(Text(s)) }
func Bytea(b []byte) pgtype.Bytea     { return pgtype.Bytea{Bytes: b, Status: pgtype.Present} }
func JSONB(b []byte) pgtype.JSONB     { return pgtype.JSONB{Bytes: b, Status: pgtype.Present} }

func Now() pgtype.Timestamp                      { return pgtype.Timestamp{Time: time.Now(), Status: pgtype.Present} }
func Timestamp(t time.Time) pgtype.Timestamp     { return pgtype.Timestamp{Time: t, Status: pgtype.Present} }
func Timestamptz(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Status: pgtype.Present} }

func Int2(i int16) pgtype.Int2  { return pgtype.Int2{Int: i, Status: pgtype.Present} }
func Int4(i int32) pgtype.Int4  { return pgtype.Int4{Int: i, Status: pgtype.Present} }
func Int8(i int64) pgtype.Int8  { return pgtype.Int8{Int: i, Status: pgtype.Present} }

func Float4(f float32) pgtype.Float4                    { return pgtype.Float4{Float: f, Status: pgtype.Present} }
func Float4Array(f []float32) (fa pgtype.Float4Array)   { fa.Set(f); return }
func Float32Slice(aa pgtype.Float4Array) (ff []float32) { aa.AssignTo(&ff); return }

func Float8(f float64) pgtype.Float8                    { return pgtype.Float8{Float: f, Status: pgtype.Present} }
func Float8Array(f []float64) (m pgtype.Float8Array)    { m.Set(f); return }
func Float64Slice(aa pgtype.Float8Array) (ff []float64) { aa.AssignTo(&ff); return }

func NewUUIDV4() pgtype.UUID               { return UUID(uuid.Must(uuid.NewV4())) }
func UUID(id uuid.UUID) pgtype.UUID        { return pgtype.UUID{Bytes: [16]byte(id), Status: pgtype.Present} }
func UUIDFromString(id string) pgtype.UUID { return pgtype.UUID{Bytes: uuid.FromStringOrNil(id), Status: pgtype.Present} }
func UUIDArray(ids []uuid.UUID) pgtype.UUIDArray {
	m := pgtype.UUIDArray{}
	b := make([][16]byte, len(ids))
	for k, e := range ids {
		b[k] = [16]byte(e)
	}
	m.Set(b)
	return m
}
func ToUUIDSlice(a pgtype.UUIDArray) []uuid.UUID {
	bs := make([]uuid.UUID, len(a.Elements))
	for k, e := range a.Elements {
		bs[k] = uuid.FromBytesOrNil(e.Bytes[:])
	}
	return bs
}
