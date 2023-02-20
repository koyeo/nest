package _range

import "testing"

func TestInt64Ranges(t *testing.T) {

	t.Log(Int64Ranges(1, 5, 2))
	t.Log(Int64Ranges(1, 8, 2))
	t.Log(Int64Ranges(0, 3, 5))
	t.Log(Int64Ranges(3, 9, 4))

}
