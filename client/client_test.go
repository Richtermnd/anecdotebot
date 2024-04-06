package client_test

import (
	"strconv"
	"testing"

	"github.com/Richtermnd/anecdotebot/client"
)

// Just test that it doesn't panic and doesn't return empty string
// !!! Make 50 requests, so it can take some time.
func TestGetAnecdote(t *testing.T) {
	t.Parallel()
	for i := 0; i < 50; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			if client.GetAnecdote(1) == "" {
				t.Fatal()
			}
		})
	}
}

func TestGetContent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "Normal",
			in:   `{"content":"test"}`,
			out:  "test",
		},
		{
			name: "With newlines",
			in:   "{\"content\":\"with\nnewlines\"}",
			out:  "with\nnewlines",
		},
		{
			name: "Non ascii symbols",
			in:   "{\"content\":\"Русский текст\"}",
			out:  "Русский текст",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			out := client.GetContent([]byte(tC.in))
			if out != tC.out {
				t.Errorf("expected %s, got %s", tC.out, out)
			}
		})
	}
}
