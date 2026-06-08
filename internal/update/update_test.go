package update

import "testing"

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.6.2", "v1.6.1", 1},
		{"v1.7.0", "v1.6.9", 1},
		{"v2.0.0", "v1.9.9", 1},
		{"v1.6.1", "v1.6.1", 0},
		{"1.6.1", "v1.6.1", 0},
		{"v1.6.0", "v1.6.1", -1},
		{"v1.6.1-rc.1", "v1.6.1", 0},
		{"v1.6.2-rc.1", "v1.6.1", 1},
	}
	for _, c := range cases {
		got := compare(c.a, c.b)
		if got != c.want {
			t.Errorf("compare(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
