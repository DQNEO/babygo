package strconv

import "testing"

func TestParseInt(t *testing.T) {
	type args struct {
		s       string
		base    int
		bitSize int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"empty is zero", args{"", 0, 64}, 0, false},
		{"digit 0", args{"0", 0, 64}, 0, false},
		{"digit 9", args{"9", 0, 64}, 9, false},
		{"digit 789", args{"789", 0, 64}, 789, false},
		{"hex a", args{"0xa", 0, 64}, 10, false},
		{"hex 78", args{"0x78", 0, 64}, 120, false},
		{"hex ff", args{"0xff", 0, 64}, 255, false},
		{"hex ffff", args{"0xffff", 0, 64}, 65535, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt(tt.args.s, tt.args.base, tt.args.bitSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseInt() got = %v, want %v", got, tt.want)
			}
		})
	}
}
