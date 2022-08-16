package utils

type Numeric interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

func Sum[T Numeric](values []T) T {
	var tmp T

	for _, val := range values {
		tmp += val
	}
	return tmp
}
