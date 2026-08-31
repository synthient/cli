package feed

import "testing"

func TestFindJA4Streams(t *testing.T) {
	tests := []string{"ja4", "ja4t"}
	for _, name := range tests {
		stream, ok := Find(name)
		if !ok {
			t.Fatalf("Find(%q) did not find stream", name)
		}
		if stream.Name != name {
			t.Fatalf("Find(%q) returned %q", name, stream.Name)
		}
	}
}

func TestNamesIncludesJA4Streams(t *testing.T) {
	want := "proxies|anonymizers|torrents|ja4|ja4t|honeypot_http|honeypot_https|honeypot_dns|honeypot_adb"
	got := Names()
	if got != want {
		t.Fatalf("Names() = %q, want %q", got, want)
	}
}
