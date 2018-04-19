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
	"encoding/hex"
	"strconv"

	"github.com/jackc/pgx/pgtype"
	"github.com/satori/go.uuid"
)

func TextStr(v pgtype.Text) string               { return v.String }
func VarcharStr(v pgtype.Text) string            { return v.String }
func ByteaStr(v pgtype.Bytea) string             { return hex.EncodeToString(v.Bytes) }
func Int2Str(v pgtype.Int2) string               { return strconv.FormatInt(int64(v.Int), 10) }
func Int4Str(v pgtype.Int4) string               { return strconv.FormatInt(int64(v.Int), 10) }
func Int8Str(v pgtype.Int8) string               { return strconv.FormatInt(v.Int, 10) }
func Float4Str(v pgtype.Float4) string           { return strconv.FormatFloat(float64(v.Float), 'f', -1, 32) }
func Float8Str(v pgtype.Float8) string           { return strconv.FormatFloat(v.Float, 'f', -1, 64) }
func BoolStr(v pgtype.Bool) string               { return strconv.FormatBool(v.Bool) }
func UUIDStr(v pgtype.UUID) string               { return uuid.FromBytesOrNil(v.Bytes[:]).String() }
func TimestampStr(v pgtype.Timestamp) string     { return v.Time.String() }
func TimestamptzStr(v pgtype.Timestamptz) string { return v.Time.String() }
