package gh_test

import (
	"fmt"
	"testing"

	"github.com/kkohtaka/gh-actions-pr-size/pkg/gh"
	"github.com/stretchr/testify/assert"
)

func TestNewSize(t *testing.T) {
	tcs := []struct {
		change     int
		wantLabel  string
		wantString string
	}{
		{
			change:     0,
			wantLabel:  "size/XS",
			wantString: "XS",
		},
		{
			change:     9,
			wantLabel:  "size/XS",
			wantString: "XS",
		},
		{
			change:     10,
			wantLabel:  "size/S",
			wantString: "S",
		},
		{
			change:     11,
			wantLabel:  "size/S",
			wantString: "S",
		},
		{
			change:     29,
			wantLabel:  "size/S",
			wantString: "S",
		},
		{
			change:     30,
			wantLabel:  "size/M",
			wantString: "M",
		},
		{
			change:     31,
			wantLabel:  "size/M",
			wantString: "M",
		},
		{
			change:     99,
			wantLabel:  "size/M",
			wantString: "M",
		},
		{
			change:     100,
			wantLabel:  "size/L",
			wantString: "L",
		},
		{
			change:     101,
			wantLabel:  "size/L",
			wantString: "L",
		},
		{
			change:     499,
			wantLabel:  "size/L",
			wantString: "L",
		},
		{
			change:     500,
			wantLabel:  "size/XL",
			wantString: "XL",
		},
		{
			change:     501,
			wantLabel:  "size/XL",
			wantString: "XL",
		},
		{
			change:     999,
			wantLabel:  "size/XL",
			wantString: "XL",
		},
		{
			change:     1000,
			wantLabel:  "size/XXL",
			wantString: "XXL",
		},
		{
			change:     1001,
			wantLabel:  "size/XXL",
			wantString: "XXL",
		},
	}
	for _, tt := range tcs {
		t.Run(fmt.Sprintf("newSize(%d) => %s / %s", tt.change, tt.wantLabel, tt.wantString), func(t *testing.T) {
			got := gh.NewSize(tt.change)
			assert.Equal(t, tt.wantLabel, got.GetLabel())
			assert.Equal(t, tt.wantString, got.String())
		})
	}
}

func TestSizeUnknown(t *testing.T) {
	var size gh.Size = 999
	assert.Equal(t, "Unknown", size.String())
	assert.Equal(t, "size/?", size.GetLabel())
}
