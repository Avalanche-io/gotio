// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentime

import (
	"testing"
)

// =============================================================================
// RationalTime Benchmarks
// =============================================================================

func BenchmarkRationalTime_Add_SameRate(b *testing.B) {
	rt1 := NewRationalTime(100, 24)
	rt2 := NewRationalTime(50, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rt1.Add(rt2)
	}
}

func BenchmarkRationalTime_Add_DifferentRates(b *testing.B) {
	rt1 := NewRationalTime(100, 24)
	rt2 := NewRationalTime(50, 30)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rt1.Add(rt2)
	}
}

func BenchmarkRationalTime_Sub(b *testing.B) {
	rt1 := NewRationalTime(100, 24)
	rt2 := NewRationalTime(50, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rt1.Sub(rt2)
	}
}

func BenchmarkRationalTime_RescaledTo(b *testing.B) {
	rt := NewRationalTime(1000, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rt.RescaledTo(30)
	}
}

func BenchmarkRationalTime_ToFrames(b *testing.B) {
	rt := NewRationalTime(1000.5, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rt.ToFrames()
	}
}

func BenchmarkRationalTime_ToSeconds(b *testing.B) {
	rt := NewRationalTime(1000, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rt.ToSeconds()
	}
}

func BenchmarkRationalTime_ToTimecode_24fps(b *testing.B) {
	rt := NewRationalTime(86400, 24) // 1 hour at 24fps
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = rt.ToTimecode(24, ForceNo)
	}
}

func BenchmarkRationalTime_ToTimecode_2997fps_DropFrame(b *testing.B) {
	rt := NewRationalTime(107892, 29.97) // ~1 hour at 29.97fps
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = rt.ToTimecode(29.97, ForceYes)
	}
}

func BenchmarkRationalTime_Cmp(b *testing.B) {
	rt1 := NewRationalTime(100, 24)
	rt2 := NewRationalTime(100.5, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rt1.Cmp(rt2)
	}
}

func BenchmarkRationalTime_AlmostEqual(b *testing.B) {
	rt1 := NewRationalTime(100, 24)
	rt2 := NewRationalTime(100.0001, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rt1.AlmostEqual(rt2, 0.001)
	}
}

// =============================================================================
// TimeRange Benchmarks
// =============================================================================

func BenchmarkTimeRange_Contains(b *testing.B) {
	tr := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	t := NewRationalTime(50, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tr.Contains(t)
	}
}

func BenchmarkTimeRange_ContainsRange(b *testing.B) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	tr2 := NewTimeRange(NewRationalTime(25, 24), NewRationalTime(50, 24))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tr1.ContainsRange(tr2, DefaultEpsilon)
	}
}

func BenchmarkTimeRange_Intersects(b *testing.B) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	tr2 := NewTimeRange(NewRationalTime(50, 24), NewRationalTime(100, 24))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tr1.Intersects(tr2, DefaultEpsilon)
	}
}

func BenchmarkTimeRange_ClampedRange(b *testing.B) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	tr2 := NewTimeRange(NewRationalTime(50, 24), NewRationalTime(200, 24))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tr1.ClampedRange(tr2)
	}
}

func BenchmarkTimeRange_ClampedTime(b *testing.B) {
	tr := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	t := NewRationalTime(150, 24) // Outside range
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tr.ClampedTime(t)
	}
}

func BenchmarkTimeRange_ExtendedBy(b *testing.B) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	tr2 := NewTimeRange(NewRationalTime(50, 24), NewRationalTime(200, 24))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tr1.ExtendedBy(tr2)
	}
}

func BenchmarkTimeRange_EndTimeExclusive(b *testing.B) {
	tr := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tr.EndTimeExclusive()
	}
}

func BenchmarkTimeRange_EndTimeInclusive(b *testing.B) {
	tr := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tr.EndTimeInclusive()
	}
}

func BenchmarkTimeRange_Equal(b *testing.B) {
	tr1 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	tr2 := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tr1.Equal(tr2)
	}
}

// =============================================================================
// TimeTransform Benchmarks
// =============================================================================

func BenchmarkTimeTransform_AppliedToTime(b *testing.B) {
	tt := NewTimeTransform(NewRationalTime(10, 24), 2.0, 24)
	rt := NewRationalTime(100, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tt.AppliedToTime(rt)
	}
}

func BenchmarkTimeTransform_AppliedToRange(b *testing.B) {
	tt := NewTimeTransform(NewRationalTime(10, 24), 2.0, 24)
	tr := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tt.AppliedToRange(tr)
	}
}

// =============================================================================
// JSON Serialization Benchmarks
// =============================================================================

func BenchmarkRationalTime_MarshalJSON(b *testing.B) {
	rt := NewRationalTime(1000, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = rt.MarshalJSON()
	}
}

func BenchmarkRationalTime_UnmarshalJSON(b *testing.B) {
	data := []byte(`{"OTIO_SCHEMA":"RationalTime.1","value":1000,"rate":24}`)
	rt := RationalTime{}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rt.UnmarshalJSON(data)
	}
}

func BenchmarkTimeRange_MarshalJSON(b *testing.B) {
	tr := NewTimeRange(NewRationalTime(0, 24), NewRationalTime(100, 24))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = tr.MarshalJSON()
	}
}

func BenchmarkTimeRange_UnmarshalJSON(b *testing.B) {
	data := []byte(`{"OTIO_SCHEMA":"TimeRange.1","start_time":{"OTIO_SCHEMA":"RationalTime.1","value":0,"rate":24},"duration":{"OTIO_SCHEMA":"RationalTime.1","value":100,"rate":24}}`)
	tr := TimeRange{}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tr.UnmarshalJSON(data)
	}
}
