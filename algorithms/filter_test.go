// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package algorithms

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

func TestFilteredCompositionKeepAll(t *testing.T) {
	timeline := gotio.NewTimeline("test", nil, nil)
	track := gotio.NewTrack("track", nil, gotio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	result := FilteredComposition(timeline, KeepFilter, nil)
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	resultTimeline, ok := result.(*gotio.Timeline)
	if !ok {
		t.Fatal("Result should be a Timeline")
	}

	if len(resultTimeline.Tracks().Children()) != 1 {
		t.Errorf("Expected 1 track, got %d", len(resultTimeline.Tracks().Children()))
	}
}

func TestFilteredCompositionPruneAll(t *testing.T) {
	timeline := gotio.NewTimeline("test", nil, nil)
	track := gotio.NewTrack("track", nil, gotio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	timeline.Tracks().AppendChild(track)

	result := FilteredComposition(timeline, PruneFilter, nil)
	if result == nil {
		t.Fatal("Result should not be nil at root level")
	}
}

func TestFilteredCompositionTypePrune(t *testing.T) {
	timeline := gotio.NewTimeline("test", nil, nil)
	track := gotio.NewTrack("track", nil, gotio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	gap := gotio.NewGap("gap", &sr, nil, nil, nil, nil)
	track.AppendChild(clip)
	track.AppendChild(gap)
	timeline.Tracks().AppendChild(track)

	// Prune gaps
	typesToPrune := []reflect.Type{reflect.TypeOf(&gotio.Gap{})}
	result := FilteredComposition(timeline, KeepFilter, typesToPrune)

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestTypeFilter(t *testing.T) {
	filter := TypeFilter(reflect.TypeOf(&gotio.Clip{}))

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	gap := gotio.NewGap("gap", &sr, nil, nil, nil, nil)

	clipResult := filter(clip)
	if len(clipResult) != 1 {
		t.Error("TypeFilter should keep clips")
	}

	gapResult := filter(gap)
	if len(gapResult) != 0 {
		t.Error("TypeFilter should prune gaps")
	}
}

func TestNameFilter(t *testing.T) {
	filter := NameFilter(func(name string) bool {
		return strings.HasPrefix(name, "keep_")
	})

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	keepClip := gotio.NewClip("keep_clip", nil, &sr, nil, nil, nil, "", nil)
	pruneClip := gotio.NewClip("prune_clip", nil, &sr, nil, nil, nil, "", nil)

	keepResult := filter(keepClip)
	if len(keepResult) != 1 {
		t.Error("NameFilter should keep clips with matching names")
	}

	pruneResult := filter(pruneClip)
	if len(pruneResult) != 0 {
		t.Error("NameFilter should prune clips with non-matching names")
	}
}

func TestFilteredWithSequenceContext(t *testing.T) {
	timeline := gotio.NewTimeline("test", nil, nil)
	track := gotio.NewTrack("track", nil, gotio.TrackKindVideo, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := gotio.NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := gotio.NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)
	clip3 := gotio.NewClip("clip3", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)
	track.AppendChild(clip3)
	timeline.Tracks().AppendChild(track)

	// Filter that keeps items that have both prev and next
	contextFilter := func(prev, current, next gotio.Composable) []gotio.Composable {
		if prev != nil && next != nil {
			return []gotio.Composable{current}
		}
		return []gotio.Composable{current} // Keep all for this test
	}

	result := FilteredWithSequenceContext(timeline, contextFilter, nil)
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestFilteredCompositionNil(t *testing.T) {
	result := FilteredComposition(nil, KeepFilter, nil)
	if result != nil {
		t.Error("Result should be nil for nil input")
	}
}

func TestFilteredCompositionStack(t *testing.T) {
	stack := gotio.NewStack("stack", nil, nil, nil, nil, nil)
	track := gotio.NewTrack("track", nil, gotio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip)
	stack.AppendChild(track)

	result := FilteredComposition(stack, KeepFilter, nil)
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	resultStack, ok := result.(*gotio.Stack)
	if !ok {
		t.Fatal("Result should be a Stack")
	}

	if len(resultStack.Children()) != 1 {
		t.Errorf("Expected 1 child, got %d", len(resultStack.Children()))
	}
}

func TestFilteredCompositionTrack(t *testing.T) {
	track := gotio.NewTrack("track", nil, gotio.TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip1 := gotio.NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := gotio.NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)
	track.AppendChild(clip1)
	track.AppendChild(clip2)

	result := FilteredComposition(track, KeepFilter, nil)
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	resultTrack, ok := result.(*gotio.Track)
	if !ok {
		t.Fatal("Result should be a Track")
	}

	if len(resultTrack.Children()) != 2 {
		t.Errorf("Expected 2 children, got %d", len(resultTrack.Children()))
	}
}

func TestFilteredCompositionSerializableCollection(t *testing.T) {
	coll := gotio.NewSerializableCollection("coll", nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := gotio.NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	coll.AppendChild(clip)

	result := FilteredComposition(coll, KeepFilter, nil)
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	resultColl, ok := result.(*gotio.SerializableCollection)
	if !ok {
		t.Fatal("Result should be a SerializableCollection")
	}

	if len(resultColl.Children()) != 1 {
		t.Errorf("Expected 1 child, got %d", len(resultColl.Children()))
	}
}
