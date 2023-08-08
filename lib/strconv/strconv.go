package strconv

func Itoa(ival int) string {
	if ival == 0 {
		return "0"
	}

	var buf = make([]uint8, 100, 100)
	var r = make([]uint8, 100, 100)

	var next int
	var right int
	var ix = 0
	var minus bool
	minus = false
	for ix = 0; ival != 0; ix = ix + 1 {
		if ival < 0 {
			ival = -1 * ival
			minus = true
			r[0] = '-'
		} else {
			next = ival / 10
			right = ival - next*10
			ival = next
			buf[ix] = uint8('0' + right)
		}
	}

	var j int
	var c uint8
	for j = 0; j < ix; j = j + 1 {
		c = buf[ix-j-1]
		if minus {
			r[j+1] = c
		} else {
			r[j] = c
		}
	}

	return string(r[0:ix])
}

func Atoi(gs string) int {
	if len(gs) == 0 {
		return 0
	}
	var n int

	var isMinus bool
	for _, b := range []uint8(gs) {
		if b == '.' {
			panic("Unexpected")
		}
		if b == '-' {
			isMinus = true
			continue
		}
		var x uint8 = b - uint8('0')
		n = n * 10
		n = n + int(x)
	}
	if isMinus {
		n = -n
	}

	return n
}

func ParseInt(s string, base int, bitSize int) (int, error) {
	if len(s) == 0 {
		return 0, nil
	}

	if len(s) > 2 {
		prefix := s[0:2]
		switch prefix {
		case "0x":
			var n int
			s2 := s[2:]
			for _, b := range []uint8(s2) {
				if b == '.' {
					panic("Unexpected")
				}
				if '0' <= b && b <= '9' {
					var x uint8 = b - uint8('0')
					n = n * 16
					n = n + int(x)
					continue
				} else if 'a' <= b && b <= 'f' {
					var x uint8 = b - uint8('a') + 10
					n = n * 16
					n = n + int(x)
					continue
				}
			}
			return n, nil
		case "0b":
			// @TODO
			return 0, nil
		default:
			i := Atoi(s)
			return i, nil
		}
	}
	i := Atoi(s)
	return i, nil
}
