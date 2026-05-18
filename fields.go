package raven

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"
)

// Field struct holds a single key-value pair for structured logging
type Field struct {
	Name         string
	Value        string
	IsJSONString bool // value needs quotes in JSON output
	IsJSONSafe   bool // value only contains safe characters, no escaping needed
}

// Fielder interface is implemented by any type that can produce a Field
type Fielder interface {
	Field() Field
}

// Fieldify converts a slice of Fielders into a slice of Fields
func Fieldify(f []Fielder) []Field {
	fields := make([]Field, len(f))
	for i := range f {
		fields[i] = f[i].Field()
	}
	return fields
}

// FieldifyAndAppend merges existing Fields with new Fielders into one slice
func FieldifyAndAppend(fields []Field, fielders []Fielder) []Field {
	var out []Field
	if len(fielders)+len(fields) > 0 {
		out = make([]Field, len(fields), len(fielders)+len(fields))
		copy(out, fields)
		for _, fielder := range fielders {
			out = append(out, fielder.Field())
		}
	}
	return out
}

// Constructor functions

func Bool(name string, value bool) FieldBool {
	return FieldBool{Name: name, Value: value}
}

func String(name string, value string) FieldString {
	return FieldString{Name: name, Value: value}
}

func Int(name string, value int) FieldInt64 {
	return FieldInt64{Name: name, Value: int64(value)}
}

func Int64(name string, value int64) FieldInt64 {
	return FieldInt64{Name: name, Value: value}
}

func Uint(name string, value uint) FieldUint64 {
	return FieldUint64{Name: name, Value: uint64(value)}
}

func Uint64(name string, value uint64) FieldUint64 {
	return FieldUint64{Name: name, Value: value}
}

func Float64(name string, value float64) FieldFloat64 {
	return FieldFloat64{Name: name, Value: value}
}

func Err(value error) FieldError {
	return FieldError{Name: "error", Value: value}
}

func Dur(name string, value time.Duration) FieldDuration {
	return FieldDuration{Name: name, Value: value}
}

func Time(name string, value time.Time) FieldTime {
	return FieldTime{Name: name, Value: value, Format: time.RFC3339}
}

func Path(path string) FieldString {
	return FieldString{Name: "path", Value: filepath.ToSlash(path)}
}

// Stringer adds any value that implements fmt.Stringer -- this is Raven's own addition
func Stringer(name string, value fmt.Stringer) FieldString {
	return FieldString{Name: name, Value: value.String()}
}

// --- Field type implementations ---

type FieldBool struct {
	Name  string
	Value bool
}

func (f FieldBool) Field() Field {
	if f.Value {
		return Field{Name: f.Name, Value: "true"}
	}
	return Field{Name: f.Name, Value: "false"}
}

type FieldString struct {
	Name  string
	Value string
}

func (f FieldString) Field() Field {
	return Field{Name: f.Name, Value: f.Value, IsJSONString: true, IsJSONSafe: false}
}

type FieldInt64 struct {
	Name  string
	Value int64
}

func (f FieldInt64) Field() Field {
	return Field{Name: f.Name, Value: strconv.FormatInt(f.Value, 10)}
}

type FieldUint64 struct {
	Name  string
	Value uint64
}

func (f FieldUint64) Field() Field {
	return Field{Name: f.Name, Value: strconv.FormatUint(f.Value, 10)}
}

type FieldFloat64 struct {
	Name  string
	Value float64
}

func (f FieldFloat64) Field() Field {
	return Field{Name: f.Name, Value: strconv.FormatFloat(f.Value, 'g', -1, 64)}
}

type FieldError struct {
	Name  string
	Value error
}

func (f FieldError) Field() Field {
	if f.Value == nil {
		return Field{Name: f.Name, Value: "null"}
	}
	return Field{Name: f.Name, Value: f.Value.Error(), IsJSONString: true}
}

type FieldDuration struct {
	Name  string
	Value time.Duration
}

func (f FieldDuration) Field() Field {
	return Field{Name: f.Name, Value: f.Value.String(), IsJSONString: true, IsJSONSafe: true}
}

type FieldTime struct {
	Name   string
	Value  time.Time
	Format string
}

func (f FieldTime) Field() Field {
	return Field{Name: f.Name, Value: f.Value.Format(f.Format), IsJSONString: true, IsJSONSafe: true}
}