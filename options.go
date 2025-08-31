package mapper

// SetOptions allows to set mapper options with a fluent pattern, so you could
// write:
//
//	var mymapper = Mapper(...).SetOptions(WithComma(','))
//
// instead of
//
//	var mymapper = Mapper()
//	mymapper.Comma = ','
func (m *mapper) SetOptions(opts ...MapperOption) *mapper {
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type MapperOption func(m *mapper)

func WithFieldMapper(fm FieldMapper) MapperOption {
	if fm == nil {
		fm = Direct
	}
	return func(m *mapper) {
		m.FieldMapper = fm
	}
}

func WithComma(comma rune) MapperOption {
	return func(m *mapper) {
		m.Comma = comma
	}
}

func WithMark(mark rune) MapperOption {
	return func(m *mapper) {
		m.Mark = mark
	}
}
