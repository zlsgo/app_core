package common

import (
	"net/http/httptest"
	"testing"

	"github.com/sohaha/zlsgo/znet"
)

func newTestContext(rawURL string) *znet.Context {
	req := httptest.NewRequest("GET", rawURL, nil)
	return znet.New().NewContext(httptest.NewRecorder(), req)
}

func TestVarPages(t *testing.T) {
	cases := []struct {
		url          string
		wantPage     int
		wantPageSize int
	}{
		{"/?page=2&pagesize=20", 2, 20},
		{"/", 1, 10},
		{"/?page=-1&pagesize=0", 1, 10},
		{"/?page=abc&pagesize=xyz", 1, 10},
	}

	for _, tc := range cases {
		page, size, err := VarPages(newTestContext(tc.url))
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.url, err)
		}
		if page != tc.wantPage || size != tc.wantPageSize {
			t.Fatalf("%s: got page=%d pagesize=%d, want page=%d pagesize=%d",
				tc.url, page, size, tc.wantPage, tc.wantPageSize)
		}
	}
}

func TestVarUID(t *testing.T) {
	c := newTestContext("/")
	if got := VarUID(c); got != "" {
		t.Fatalf("missing uid should be empty, got %q", got)
	}

	c.WithValue("uid", 42)
	if got := VarUID(c); got != "42" {
		t.Fatalf("got %q, want %q", got, "42")
	}
}
