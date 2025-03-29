package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateOnly_MarshalUnmarshalJSON(t *testing.T) {
	original := NewDateOnly(time.Date(2025, 3, 29, 14, 35, 0, 0, time.UTC))

	data, err := json.Marshal(original)
	assert.NoError(t, err)
	assert.Equal(t, `"2025-03-29"`, string(data))

	var parsed DateOnly
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)

	assert.Equal(t, original.ToTime(), parsed.ToTime())
}

func TestDateOnly_MarshalUnmarshalJSONWithEmptyTime(t *testing.T) {
	original := NewDateOnly(time.Date(2025, 3, 29, 0, 0, 0, 0, time.UTC))

	data, err := json.Marshal(original)
	assert.NoError(t, err)
	assert.Equal(t, `"2025-03-29"`, string(data))

	var parsed DateOnly
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)

	assert.Equal(t, original.ToTime(), parsed.ToTime())
}

func TestDateOnly_ZeroDateJSON(t *testing.T) {
	var d DateOnly

	data, err := json.Marshal(d)
	assert.NoError(t, err)
	assert.Equal(t, `null`, string(data))

	var parsed DateOnly
	err = json.Unmarshal([]byte(`null`), &parsed)
	assert.NoError(t, err)
	assert.True(t, parsed.ToTime().IsZero())
}

func TestDateOnly_ValueAndScan(t *testing.T) {
	date := time.Date(2025, 3, 29, 10, 0, 0, 0, time.UTC)
	d := NewDateOnly(date)

	val, err := d.Value()
	assert.NoError(t, err)
	assert.Equal(t, date.Truncate(24*time.Hour), val.(time.Time).Truncate(24*time.Hour))

	var scanned DateOnly
	err = scanned.Scan(val)
	assert.NoError(t, err)
	assert.Equal(t, d.ToTime(), scanned.ToTime())
}
