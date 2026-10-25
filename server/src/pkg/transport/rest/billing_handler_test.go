package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_localeFromAcceptLanguage(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{name: "empty", header: "", want: ""},
		{name: "simple", header: "de", want: "de"},
		{name: "with region and weights", header: "de-DE,de;q=0.9,en;q=0.8", want: "de-DE"},
		{name: "wildcard only", header: "*", want: ""},
		{name: "spaces", header: "  en-US , en;q=0.5", want: "en-US"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, localeFromAcceptLanguage(tt.header))
		})
	}
}
