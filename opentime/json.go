// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentime

import (
	"encoding/json"
)

// rationalTimeJSON is the JSON representation of RationalTime.
type rationalTimeJSON struct {
	Schema string  `json:"OTIO_SCHEMA"`
	Rate   float64 `json:"rate"`
	Value  float64 `json:"value"`
}

// MarshalJSON implements json.Marshaler for RationalTime.
func (rt RationalTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(&rationalTimeJSON{
		Schema: "RationalTime.1",
		Rate:   rt.rate,
		Value:  rt.value,
	})
}

// UnmarshalJSON implements json.Unmarshaler for RationalTime.
func (rt *RationalTime) UnmarshalJSON(data []byte) error {
	var j rationalTimeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	rt.value = j.Value
	rt.rate = j.Rate
	return nil
}

// timeRangeJSON is the JSON representation of TimeRange.
type timeRangeJSON struct {
	Schema    string       `json:"OTIO_SCHEMA"`
	StartTime RationalTime `json:"start_time"`
	Duration  RationalTime `json:"duration"`
}

// MarshalJSON implements json.Marshaler for TimeRange.
func (tr TimeRange) MarshalJSON() ([]byte, error) {
	return json.Marshal(&timeRangeJSON{
		Schema:    "TimeRange.1",
		StartTime: tr.startTime,
		Duration:  tr.duration,
	})
}

// UnmarshalJSON implements json.Unmarshaler for TimeRange.
func (tr *TimeRange) UnmarshalJSON(data []byte) error {
	var j timeRangeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	tr.startTime = j.StartTime
	tr.duration = j.Duration
	return nil
}

// timeTransformJSON is the JSON representation of TimeTransform.
type timeTransformJSON struct {
	Schema string       `json:"OTIO_SCHEMA"`
	Offset RationalTime `json:"offset"`
	Scale  float64      `json:"scale"`
	Rate   float64      `json:"rate"`
}

// MarshalJSON implements json.Marshaler for TimeTransform.
func (tt TimeTransform) MarshalJSON() ([]byte, error) {
	return json.Marshal(&timeTransformJSON{
		Schema: "TimeTransform.1",
		Offset: tt.offset,
		Scale:  tt.scale,
		Rate:   tt.rate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for TimeTransform.
func (tt *TimeTransform) UnmarshalJSON(data []byte) error {
	var j timeTransformJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	tt.offset = j.Offset
	tt.scale = j.Scale
	tt.rate = j.Rate
	return nil
}
