package _range

import "fmt"

func Int64Ranges(a, b, step int64) (ranges [][]int64, err error) {

	if step <= 0 {
		err = fmt.Errorf("expect step > 0,got: step=%d", step)
		return
	}

	if b < a {
		err = fmt.Errorf("expect b >= a,got: a=%d, b=%d", a, b)
		return
	}

	if a == b {
		ranges = append(ranges, []int64{a, b})
		return
	}

	i := int64(0)
	for {
		aa := a + i*step
		bb := a + (i+1)*step
		if bb < b && aa < b {
			ranges = append(ranges, []int64{aa, bb})
		} else if bb >= b && aa < b {
			ranges = append(ranges, []int64{aa, b})
			break
		} else if aa >= b {
			ranges = append(ranges, []int64{aa - step, b})
			break
		}
		i++

	}

	return
}
