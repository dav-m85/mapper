// Package mapper provides a way of mapping a sql/database Scan-like output
// to a struct, and some helper functions to write them.
//
// When manipulating raw queries, it's easy to introduce bug by misordering fields,
// forgetting about names.
package mapper

// License MIT
// Author github.com/dav-m85

import (
	"reflect"
	"strings"
)

type FieldMapper func(field string) string

var Direct FieldMapper = func(field string) string { return field }

// mapper carries mapping between database columns' name and go types.
type mapper struct {
	fields []int
	cols   []string
	target reflect.Type

	// Comma is the field delimiter.
	// It is set to comma (',') by Mapper
	// Comma must be a valid rune and must not be \r, \n,
	// or the Unicode replacement character (0xFFFD).
	Comma rune

	// Mark is the field placeholder.
	// It is set to question mark ('?') by Mapper
	// Mark must be a valid rune and must not be \r, \n,
	// or the Unicode replacement character (0xFFFD).
	Mark rune

	// FieldMapper processes struct's field names when no struct tag is given.
	// It defaults to [Direct]. Common option are [strings.ToLower], [strings.ToUpper]...
	FieldMapper FieldMapper
}

// Mapper maps columns from target fields, and provides helper functions around them.
// It does field resolution in this call: consider putting it early in your
// runtime to fail fast if some columns are mistyped.
// target fields are first transformed with FieldMapper, then compared against
// the provided columns.
//
// You can change Comma and Sep directly before calling KeysString.
//
// target's fields are scanned for tags:
//
//	type A struct {
//	  Field string `mapper:"column_name"`
//	}
//
// You can change Comma, Mark, FieldMapper after instanciation with direct access or
// [SetOptions].
func Mapper(target any, columns ...string) *mapper {
	return MapperWithKey(target, "mapper", columns...)
}

// MapperWithKey changes the default tag.
func MapperWithKey(target any, key string, columns ...string) *mapper {
	if key == "" {
		panic("Mapper MUST have a non empty struct tag key.")
	}
	t := reflect.TypeOf(target)
	k := t.Kind()
	if k == reflect.Pointer {
		k = t.Elem().Kind()
	}
	if k != reflect.Struct {
		panic("Mapper first argument MUST be a struct or a struct pointer")
	}
	m := &mapper{
		Comma:       ',',
		Mark:        '?',
		FieldMapper: strings.ToLower,
		cols:        make([]string, 0, len(columns)),
		fields:      make([]int, 0, len(columns)),
		target:      reflect.TypeOf(target),
	}
	if len(columns) == 0 {
		panic("Mapper MUST select at least one field")
	}
	joker := fieldSlice(columns).joker()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.IsExported() {
			// Transform field Name to a column name
			// Check first if we have a tag for this field
			var col string
			if t := f.Tag.Get(key); t != "" {
				if strings.HasSuffix(t, ",ignore") {
					// TODO maybe add panic if this column is in columns
					continue
				}
				col = t
			} else if m.FieldMapper != nil {
				col = m.FieldMapper(f.Name)
			} else {
				col = f.Name
			}

			// Check if col is listed in wanted fields
			if !joker {
				i := fieldSlice(columns).index(col)
				if i == -1 {
					continue
				}
				// https://github.com/golang/go/wiki/SliceTricks#delete-without-preserving-order
				// Note we modify fields to exclude fields already mapped
				// I ain't empty at the end, we tried selecting things that does not exist
				columns[i] = columns[len(columns)-1]
				columns = columns[:len(columns)-1]
			}

			if fieldSlice(m.cols).index(col) != -1 {
				panic("Field " + col + " is mapped more than once")
			}

			m.cols = append(m.cols, col)
			m.fields = append(m.fields, i)
		}
	}

	if !joker && len(columns) != 0 {
		panic("Some fields are missing from target: " + strings.Join(columns, ","))
	}

	return m
}

func (m *mapper) Columns() []string {
	return m.cols
}

// Subset returns a new Mapper with only the provided columns.
//
// Columns order of original Mapper is kept.
func (m *Mapper) Subset(columns ...string) *Mapper {
	nm := &Mapper{
		Comma:       m.Comma,
		Mark:        m.Mark,
		FieldMapper: m.FieldMapper,
		cols:        make([]string, 0, len(columns)),
		fields:      make([]int, 0, len(columns)),
		target:      m.target,
	}
	if len(columns) == 0 {
		return nm
	}
	if fieldSlice(columns).joker() {
		return m
	}
	for i, c := range m.cols {
		if fieldSlice(columns).index(c) != -1 {
			nm.cols = append(nm.cols, c)
			nm.fields = append(nm.fields, m.fields[i])
		}
	}
	return nm
}

// ColumnsString return a string suitable to be used in a Select query, in the form
// column1,column2,column3
// If you need to prefix those columns, use [ColumnsStringPrefix] instead.
func (m *mapper) ColumnsString() string {
	if len(m.cols) == 1 {
		return m.cols[0]
	}

	n := len(m.cols) - 1 // one rune per Comma
	for i := 0; i < len(m.cols); i++ {
		n += len(m.cols[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(m.cols[0])
	for _, s := range m.cols[1:] {
		b.WriteRune(m.Comma)
		b.WriteString(s)
	}
	return b.String()
}

func (m *mapper) ColumnsStringPrefix(prefix string) string {
	if len(m.cols) == 1 {
		return prefix + m.cols[0]
	}

	n := len(m.cols) - 1 // one rune per Comma
	for i := 0; i < len(m.cols); i++ {
		n += len(m.cols[i]) + len(prefix)
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(prefix + m.cols[0])
	for _, s := range m.cols[1:] {
		b.WriteRune(m.Comma)
		b.WriteString(prefix + s)
	}
	return b.String()
}

// Addrs returns all mapped fields of dest as slice of addressable interfaces.
// dest must be a struct pointer.
// So if dest := &struct{a int}, Addrs will return [{*int}].
// TODO(dmo) dest == nil reuses Mapper first argument
func (m *mapper) Addrs(dest any) (res []any) {
	v := reflect.ValueOf(dest)
	if v.Type().Kind() != reflect.Pointer {
		panic("destination not a pointer")
	}
	v = v.Elem()
	if v.Type().Kind() != reflect.Struct {
		panic("destination not a struct pointer")
	}
	// TODO(dmo) check that dest same type as Mapper first argument
	for _, i := range m.fields {
		res = append(res, v.Field(i).Addr().Interface())
	}
	return
}

// Values of dest as a slice of interfaces. dest MUST be a struct or a pointer
// to a struct.
func (m *mapper) Values(dest any) (res []any) {
	v := reflect.ValueOf(dest) // dest ?
	if reflect.TypeOf(dest).Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if reflect.TypeOf(v).Kind() != reflect.Struct {
		panic("destination not a struct")
	}
	for _, i := range m.fields {
		res = append(res, v.Field(i).Interface())
	}
	return
}

// Marks returns a string of n Mark separated by Comma, where n is number of
// mapped fields.
// So then Mapper(T, "a", "b").Marks() = "?,?"
func (m *mapper) Marks() string {
	if len(m.cols) == 1 {
		return string(m.Mark)
	}

	n := 2*len(m.cols) - 1 // one rune per Comma, one per Mark

	var b strings.Builder
	b.Grow(n)
	b.WriteRune(m.Mark)
	for i := 0; i < len(m.cols)-1; i++ {
		b.WriteRune(m.Comma)
		b.WriteRune(m.Mark)
	}
	return b.String()
}

// fieldSlice helper
type fieldSlice []string

// joker is in fs
func (fs fieldSlice) joker() bool {
	for _, f := range fs {
		if f == "*" {
			return true
		}
	}
	return false
}

// index of name in fs
func (fs fieldSlice) index(name string) int {
	for i, f := range fs {
		if f == name {
			return i
		}
	}
	return -1
}
