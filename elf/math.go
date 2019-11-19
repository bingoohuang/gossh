package elf

// MaxInt returns max of integers
func MaxInt(i ...int) int {
	m := 0

	for _, ii := range i {
		if ii > m {
			m = ii
		}
	}

	return m
}
