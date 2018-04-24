// Code generated by pgxgen. DO NOT EDIT.
package {{.PackageName}}

import (
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/satori/go.uuid"
)

func String(from pgtype.Text) (r string)     { from.AssignTo(&r); return }
func String2(from pgtype.Varchar) (r string) { from.AssignTo(&r); return }

func Bool(from pgtype.Bool) (r bool)     { from.AssignTo(&r); return }
func Bytes(from pgtype.Bytea) (r []byte) { from.AssignTo(&r); return }
func JSONB(from pgtype.JSONB) (r []byte) { from.AssignTo(&r); return }

func Time(from pgtype.Timestamp) (r time.Time)    { from.AssignTo(&r); return }
func Time2(from pgtype.Timestamptz) (r time.Time) { from.AssignTo(&r); return }

func Int16(from pgtype.Int2) (r int16) { from.AssignTo(&r); return }
func Int32(from pgtype.Int4) (r int32) { from.AssignTo(&r); return }
func Int64(from pgtype.Int8) (r int64) { from.AssignTo(&r); return }

func Float32(from pgtype.Float4) (r float32)             { from.AssignTo(&r); return }
func Float32Slice(from pgtype.Float4Array) (r []float32) { from.AssignTo(&r); return }

func Float64(from pgtype.Float8) (r float64)             { from.AssignTo(&r); return }
func Float64Slice(from pgtype.Float8Array) (r []float64) { from.AssignTo(&r); return }

func UUID(id pgtype.UUID) uuid.UUID {
	var b []byte
	id.AssignTo(&b)
	return uuid.FromBytesOrNil(b[:])
}
func UUIDSlice(aa pgtype.UUIDArray) []uuid.UUID {
	bs := make([]uuid.UUID, len(aa.Elements))
	for k, e := range aa.Elements {
		bs[k] = uuid.FromBytesOrNil(e.Bytes[:])
	}
	return bs
}