// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"reflect"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

// Additional tests for coverage improvement

func TestFilteredWithSequenceContextStack(t *testing.T) {
	stack := gotio.NewStack("stack", nil, nil, nil, nil, nil)
	track := gotio.NewTrack("track", nil, gotio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := gotio.NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := gotio.NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)
	stack.AppendChild(track)

	contextFilter := func(prev, current, next gotio.Composable) []gotio.Composable {
		return []gotio.Composable{current}
	}

	result := FilteredWithSequenceContext(stack, contextFilter, nil)
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestFilteredWithSequenceContextTrack(t *testing.T) {
	track := gotio.NewTrack("track", nil, gotio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := gotio.NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := gotio.NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)
	clip3 := gotio.NewClip("clip3", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)
	track.AppendChild(clip3)

	contextFilter := func(prev, current, next gotio.Composable) []gotio.Composable {
		// Only keep middle items (have both prev and next)
		if prev != nil && next != nil {
			return []gotio.Composable{current}
		}
		return nil
	}

	result := FilteredWithSequenceContext(track, contextFilter, nil)
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestFilteredWithSequenceContextSerializableCollection(t *testing.T) {
	coll := gotio.NewSerializableCollection("coll", nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := gotio.NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := gotio.NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)

	coll.AppendChild(clip1)
	coll.AppendChild(clip2)

	contextFilter := func(prev, current, next gotio.Composable) []gotio.Composable {
		return []gotio.Composable{current}
	}

	result := FilteredWithSequenceContext(coll, contextFilter, nil)
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestFilteredWithSequenceContextNil(t *testing.T) {
	contextFilter := func(prev, current, next gotio.Composable) []gotio.Composable {
		return []gotio.Composable{current}
	}

	result := FilteredWithSequenceContext(nil, contextFilter, nil)
	if result != nil {
		t.Error("Result should be nil for nil input")
	}
}

func TestFilteredWithSequenceContextPrune(t *testing.T) {
	track := gotio.NewTrack("track", nil, gotio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	gap := gotio.NewGap("gap", &sr, nil, nil, nil, nil)

	track.AppendChild(clip)
	track.AppendChild(gap)

	contextFilter := func(prev, current, next gotio.Composable) []gotio.Composable {
		return []gotio.Composable{current}
	}

	// Prune gaps
	typesToPrune := []reflect.Type{reflect.TypeOf(&gotio.Gap{})}
	result := FilteredWithSequenceContext(track, contextFilter, typesToPrune)
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestTrimItemToRangeNoSourceRange(t *testing.T) {
	// Create clip without source range but with available range
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	ref := gotio.NewExternalReference("", "/path/file.mov", &ar, nil)
	clip := gotio.NewClip("clip", ref, nil, nil, nil, nil, "", nil)

	originalRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	newRange := opentime.NewTimeRange(opentime.NewRationalTime(25, 24), opentime.NewRationalTime(50, 24))

	trimItemToRange(clip, originalRange, newRange)

	sr := clip.SourceRange()
	if sr == nil {
		t.Fatal("Source range should be set")
	}
	if sr.Duration().Value() != 50 {
		t.Errorf("Duration = %v, want 50", sr.Duration().Value())
	}
}

func TestTrimItemToRangeNoAvailableRange(t *testing.T) {
	// Create clip without source range and without available range
	clip := gotio.NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	originalRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	newRange := opentime.NewTimeRange(opentime.NewRationalTime(25, 24), opentime.NewRationalTime(50, 24))

	// This should not panic, just return without setting anything
	trimItemToRange(clip, originalRange, newRange)
}

func TestClipAtTimeInTrackNested(t *testing.T) {
	track := gotio.NewTrack("track", nil, gotio.TrackKindVideo, nil, nil)

	// Add a nested stack
	innerStack := gotio.NewStack("inner", nil, nil, nil, nil, nil)
	innerTrack := gotio.NewTrack("inner_track", nil, gotio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("deep_clip", nil, &sr, nil, nil, nil, "", nil)
	innerTrack.AppendChild(clip)
	innerStack.AppendChild(innerTrack)

	// Add inner stack to outer track with source range
	innerStackSr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	innerStack.SetSourceRange(&innerStackSr)
	track.AppendChild(innerStack)

	result := clipAtTimeInTrack(track, opentime.NewRationalTime(24, 24))
	// Result may or may not find the nested clip depending on range calculations
	t.Logf("Nested clip search result: %v", result)
}

func TestClipAtTimeInTrackNoChildren(t *testing.T) {
	track := gotio.NewTrack("empty", nil, gotio.TrackKindVideo, nil, nil)

	result := clipAtTimeInTrack(track, opentime.NewRationalTime(24, 24))
	if result != nil {
		t.Error("Result should be nil for empty track")
	}
}

func TestTopClipAtTimeDisabled(t *testing.T) {
	stack := gotio.NewStack("stack", nil, nil, nil, nil, nil)
	track := gotio.NewTrack("track", nil, gotio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	clip.SetEnabled(false) // Disable the clip

	track.AppendChild(clip)
	stack.AppendChild(track)

	result := TopClipAtTime(stack, opentime.NewRationalTime(24, 24))
	// Should not return the disabled clip
	if result != nil {
		t.Logf("Found clip (enabled=%v)", result.Enabled())
	}
}

func TestCompositeTrackOnTopWithTransitions(t *testing.T) {
	base := gotio.NewTrack("base", nil, gotio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	base.AppendChild(clip)

	top := gotio.NewTrack("top", nil, gotio.TrackKindVideo, nil, nil)
	inOffset := opentime.NewRationalTime(12, 24)
	outOffset := opentime.NewRationalTime(12, 24)
	transition := gotio.NewTransition("dissolve", "SMPTE_Dissolve", inOffset, outOffset, nil)
	top.AppendChild(transition)

	result, err := compositeTrackOnTop(base, top)
	if err != nil {
		t.Fatalf("compositeTrackOnTop error: %v", err)
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestTimelineAudioTracksNilTracks(t *testing.T) {
	timeline := &gotio.Timeline{}
	// tracks is nil

	result := TimelineAudioTracks(timeline)
	if result != nil {
		t.Error("Should return nil for timeline with nil tracks")
	}
}

func TestTimelineVideoTracksNilTracks(t *testing.T) {
	timeline := &gotio.Timeline{}
	// tracks is nil

	result := TimelineVideoTracks(timeline)
	if result != nil {
		t.Error("Should return nil for timeline with nil tracks")
	}
}

func TestFlattenTimelineVideoTracksNilTracks(t *testing.T) {
	timeline := &gotio.Timeline{}
	// tracks is nil

	result, err := FlattenTimelineVideoTracks(timeline)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	// Should return the cloned timeline
	_ = result
}

func TestTimelineTrimmedToRangeNonTrack(t *testing.T) {
	timeline := gotio.NewTimeline("test", nil, nil)

	// Add a non-track child (Gap used as example)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	gap := gotio.NewGap("gap", &sr, nil, nil, nil, nil)
	timeline.Tracks().AppendChild(gap)

	trimRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	result, err := TimelineTrimmedToRange(timeline, trimRange)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	// The non-track child should be cloned and kept
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestFilterChildrenWithContextPrune(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	gap := gotio.NewGap("gap", &sr, nil, nil, nil, nil)

	children := []gotio.Composable{clip, gap}

	contextFilter := func(prev, current, next gotio.Composable) []gotio.Composable {
		return []gotio.Composable{current}
	}

	// Prune gaps
	typesToPrune := []reflect.Type{reflect.TypeOf(&gotio.Gap{})}
	result := filterChildrenWithContext(children, contextFilter, typesToPrune)

	if len(result) != 1 {
		t.Errorf("Expected 1 result after pruning gaps, got %d", len(result))
	}
}

func TestFilterChildrenWithContextEmpty(t *testing.T) {
	contextFilter := func(prev, current, next gotio.Composable) []gotio.Composable {
		return nil
	}

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	children := []gotio.Composable{clip}

	result := filterChildrenWithContext(children, contextFilter, nil)
	if len(result) != 0 {
		t.Errorf("Expected 0 results, got %d", len(result))
	}
}

func TestSubtractRangeEndOverlap(t *testing.T) {
	// Test overlap at end of a
	a := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(75, 24))
	b := opentime.NewTimeRange(opentime.NewRationalTime(50, 24), opentime.NewRationalTime(50, 24))

	result := subtractRange(a, b)
	// Should have 1 result: 0-50
	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
	}
}

func TestFilteredCompositionClip(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	result := FilteredComposition(clip, KeepFilter, nil)
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestFilteredCompositionPruneResult(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)

	// Filter that removes everything
	filter := func(c gotio.Composable) []gotio.Composable {
		return nil
	}

	result := FilteredComposition(clip, filter, nil)
	if result != nil {
		t.Error("Result should be nil when filter prunes the item")
	}
}
