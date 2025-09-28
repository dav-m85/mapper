package mapper

import (
	"testing"

	"github.com/matryer/is"
)

func TestKeys(t *testing.T) {
	is := is.New(t)
	type M struct {
		A string
		b int32
	}

	res := Mapper(M{}, "*").Columns()
	is.Equal(res, []string{"a"})
}

func TestValues(t *testing.T) {
	is := is.New(t)
	type M struct {
		A string
		b int32
	}

	m := new(M)
	vs := Mapper(M{}, "*").Addrs(m)
	a := vs[0]
	as, ok := a.(*string)
	is.True(ok)
	*as = "yolo"
	is.Equal(m.A, "yolo")
}

func TestTag(t *testing.T) {
	is := is.New(t)
	type M struct {
		A string `mapper:"b"`
	}

	dut := Mapper(M{}, "*")
	is.Equal(dut.Columns(), []string{"b"})
}

func TestDefaultFieldMapper(t *testing.T) {
	is := is.New(t)
	type M struct {
		A  string
		Ab string
	}

	dut := Mapper(M{}, "*")
	is.Equal(dut.Columns(), []string{"a", "ab"})
}

func TestMapperDuplicate(t *testing.T) {
	is := is.New(t)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		} else {
			is.Equal(r.(string), "Field b is mapped more than once")
		}
	}()
	type M struct {
		A string `mapper:"b"`
		B string
	}

	Mapper(M{}, "*")
}

func TestMapperWithKey(t *testing.T) {
	is := is.New(t)

	type M struct {
		A        string `blah:"d_84"`
		B        string `blah:"b"`
		HomeAway string
		GoodBye  string `blah:"good_bye"`
		D        string `blah:"1"`
		E        string `blah:"42"`
	}

	mapper := MapperWithKey(M{}, "blah", "*")

	is.Equal(mapper.Columns(), []string{"d_84", "b", "homeaway", "good_bye", "1", "42"})
}

func TestMapperStructSubset(t *testing.T) {
	is := is.New(t)
	type M struct {
		A string
		B string
	}

	dut := Mapper(M{}, "a")
	is.Equal(dut.Columns(), []string{"a"})
}

func TestMapperUnknownField(t *testing.T) {
	is := is.New(t)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		} else {
			is.Equal(r.(string), "Some fields are missing from target: c")
		}
	}()
	type M struct {
		A string
		B string
	}

	Mapper(M{}, "a", "c")
}

func TestMapperSubset(t *testing.T) {
	is := is.New(t)
	type M struct {
		A string
		B string
		C string
		D string
	}

	dut := New(M{}, "*")
	is.Equal(dut.Columns(), []string{"a", "b", "c", "d"})

	sub := dut.Subset("a", "c")
	is.Equal(sub.Columns(), []string{"a", "c"})

	sub2 := dut.Subset("c", "a")
	is.Equal(sub2.Columns(), []string{"a", "c"})

	sub3 := dut.Subset("b")
	is.Equal(sub3.Columns(), []string{"b"})

	sub4 := dut.Subset()
	is.Equal(sub4.Columns(), []string{})

	sub5 := dut.Subset("*")
	is.Equal(sub5.Columns(), []string{"a", "b", "c", "d"})
}
