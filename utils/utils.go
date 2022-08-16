package utils

func Sum_floats64(values []float64) float64 {
	var tmp float64

	for _, val := range values {
		tmp = tmp + val
	}
	return tmp
}

func Sum_uint64(values []uint64) uint64 {
	var tmp uint64

	for _, val := range values {
		tmp = tmp + val
	}
	return tmp
}

func Sum_int(values []int) int {
	var tmp int

	for _, val := range values {
		tmp = tmp + val
	}
	return tmp
}
