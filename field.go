package errors

type field struct {
	key   string
	value interface{}
}

func (f field) Value() interface{} {
	return f.value
}

func (f field) Key() string {
	return f.key
}

func Bool(key string, b bool) field {
	return field{key, b}
}

func Float32(key string, f float32) field {
	return field{key, f}
}

func Float64(key string, f float64) field {
	return field{key, f}
}

func Int(key string, i int) field {
	return field{key, i}
}

func Int64(key string, i int64) field {
	return field{key, i}
}

func String(key string, s string) field {
	return field{key, s}
}
