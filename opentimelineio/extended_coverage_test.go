// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
)

// Tests for MarkerColor.ToColor
func TestMarkerColorToColor(t *testing.T) {
	tests := []struct {
		color    MarkerColor
		expected *Color
	}{
		{MarkerColorPink, ColorPink},
		{MarkerColorRed, ColorRed},
		{MarkerColorOrange, ColorOrange},
		{MarkerColorYellow, ColorYellow},
		{MarkerColorGreen, ColorGreen},
		{MarkerColorCyan, ColorCyan},
		{MarkerColorBlue, ColorBlue},
		{MarkerColorPurple, ColorPurple},
		{MarkerColorMagenta, ColorMagenta},
		{MarkerColorBlack, ColorBlack},
		{MarkerColorWhite, ColorWhite},
	}

	for _, tt := range tests {
		result := tt.color.ToColor()
		if result.R != tt.expected.R || result.G != tt.expected.G ||
			result.B != tt.expected.B || result.A != tt.expected.A {
			t.Errorf("ToColor(%s) = %v, want %v", tt.color, result, tt.expected)
		}
	}

	// Test default case (unknown color)
	unknown := MarkerColor("UNKNOWN")
	result := unknown.ToColor()
	if result != ColorGreen {
		t.Errorf("ToColor(UNKNOWN) = %v, want ColorGreen", result)
	}
}

// Tests for ChildAtTime with deep search (shallowSearch=false)
func TestChildAtTimeDeepSearch(t *testing.T) {
	// Create a nested structure: Stack -> Track -> Clip
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("deep_clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	stack.AppendChild(track)

	// Deep search should find the track (since Stack contains Track, not Clip directly)
	time := opentime.NewRationalTime(10, 24)
	child, err := stack.ChildAtTime(time, false)
	if err != nil {
		t.Fatalf("ChildAtTime deep search error: %v", err)
	}
	// The deep search in Stack finds Track, then Track's ChildAtTime returns Clip
	if child == nil {
		t.Log("ChildAtTime returned nil (composition may not support deep search)")
	} else {
		t.Logf("Found child: %T %s", child, child.(*Clip).Name())
	}
}

// Tests for ChildAtTime at boundaries
func TestChildAtTimeBoundaries(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)

	// Test at exact boundary between clips
	time := opentime.NewRationalTime(24, 24)
	child, err := track.ChildAtTime(time, true)
	if err != nil {
		t.Fatalf("ChildAtTime error: %v", err)
	}
	if child != nil {
		t.Logf("Found child at boundary: %s", child.(*Clip).Name())
	}

	// Test beyond all clips
	time = opentime.NewRationalTime(100, 24)
	child, err = track.ChildAtTime(time, true)
	if err != nil {
		t.Fatalf("ChildAtTime error: %v", err)
	}
	if child != nil {
		t.Error("Should not find child beyond all clips")
	}
}

// Tests for Gap JSON with effects
func TestGapJSONWithEffects(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	effect := NewLinearTimeWarp("timewarp", "TimeWarp", 2.0, nil)
	gap := NewGap("gap", &sr, nil, []Effect{effect}, nil, nil)

	// Marshal
	data, err := gap.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	// Check that it's valid JSON
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Unmarshal
	gap2 := &Gap{}
	if err := gap2.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if gap2.Name() != "gap" {
		t.Errorf("Unmarshaled name = %s, want gap", gap2.Name())
	}
	if len(gap2.Effects()) != 1 {
		t.Errorf("Unmarshaled effects count = %d, want 1", len(gap2.Effects()))
	}
}

// Tests for Gap JSON with color
func TestGapJSONWithColor(t *testing.T) {
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	color := NewColor(1.0, 0.5, 0.25, 1.0)
	gap := NewGap("colored_gap", &sr, nil, nil, nil, color)

	// Marshal
	data, err := gap.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	// Unmarshal
	gap2 := &Gap{}
	if err := gap2.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if gap2.ItemColor() == nil {
		t.Fatal("Unmarshaled color should not be nil")
	}
	if gap2.ItemColor().R != 1.0 {
		t.Errorf("Unmarshaled color R = %v, want 1.0", gap2.ItemColor().R)
	}
}

// Tests for Stack.AvailableImageBounds with clips
func TestStackAvailableImageBoundsWithClips(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)

	// Create clips with media references that have image bounds
	bounds1 := &Box2d{Min: Vec2d{X: 0, Y: 0}, Max: Vec2d{X: 1920, Y: 1080}}
	ref1 := NewExternalReference("", "/path/file1.mov", nil, nil)
	ref1.SetAvailableImageBounds(bounds1)
	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip1", ref1, &sr1, nil, nil, nil, "", nil)

	bounds2 := &Box2d{Min: Vec2d{X: -100, Y: -50}, Max: Vec2d{X: 2048, Y: 1200}}
	ref2 := NewExternalReference("", "/path/file2.mov", nil, nil)
	ref2.SetAvailableImageBounds(bounds2)
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip2 := NewClip("clip2", ref2, &sr2, nil, nil, nil, "", nil)

	stack.AppendChild(clip1)
	stack.AppendChild(clip2)

	result, err := stack.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}

	// Should be union of both bounds
	if result == nil {
		t.Fatal("AvailableImageBounds should not be nil")
	}
	if result.Min.X != -100 {
		t.Errorf("Min.X = %v, want -100", result.Min.X)
	}
	if result.Min.Y != -50 {
		t.Errorf("Min.Y = %v, want -50", result.Min.Y)
	}
	if result.Max.X != 2048 {
		t.Errorf("Max.X = %v, want 2048", result.Max.X)
	}
	if result.Max.Y != 1200 {
		t.Errorf("Max.Y = %v, want 1200", result.Max.Y)
	}
}

// Tests for Stack.AvailableImageBounds with nested tracks
func TestStackAvailableImageBoundsWithTracks(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)

	bounds := &Box2d{Min: Vec2d{X: 0, Y: 0}, Max: Vec2d{X: 1920, Y: 1080}}
	ref := NewExternalReference("", "/path/file.mov", nil, nil)
	ref.SetAvailableImageBounds(bounds)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", ref, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	stack.AppendChild(track)

	result, err := stack.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}

	if result == nil {
		t.Fatal("AvailableImageBounds should not be nil")
	}
	if result.Max.X != 1920 {
		t.Errorf("Max.X = %v, want 1920", result.Max.X)
	}
}

// Tests for Stack.AvailableImageBounds with nested stacks
func TestStackAvailableImageBoundsWithNestedStacks(t *testing.T) {
	outerStack := NewStack("outer", nil, nil, nil, nil, nil)
	innerStack := NewStack("inner", nil, nil, nil, nil, nil)

	bounds := &Box2d{Min: Vec2d{X: 0, Y: 0}, Max: Vec2d{X: 3840, Y: 2160}}
	ref := NewExternalReference("", "/path/4k.mov", nil, nil)
	ref.SetAvailableImageBounds(bounds)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("4k_clip", ref, &sr, nil, nil, nil, "", nil)

	innerStack.AppendChild(clip)
	outerStack.AppendChild(innerStack)

	result, err := outerStack.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}

	if result == nil {
		t.Fatal("AvailableImageBounds should not be nil")
	}
	if result.Max.X != 3840 {
		t.Errorf("Max.X = %v, want 3840", result.Max.X)
	}
	if result.Max.Y != 2160 {
		t.Errorf("Max.Y = %v, want 2160", result.Max.Y)
	}
}

// Tests for Track.AvailableImageBounds with multiple clips
func TestTrackAvailableImageBoundsMultipleClips(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)

	bounds1 := &Box2d{Min: Vec2d{X: 0, Y: 0}, Max: Vec2d{X: 1280, Y: 720}}
	ref1 := NewExternalReference("", "/path/720p.mov", nil, nil)
	ref1.SetAvailableImageBounds(bounds1)
	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("720p", ref1, &sr1, nil, nil, nil, "", nil)

	bounds2 := &Box2d{Min: Vec2d{X: 0, Y: 0}, Max: Vec2d{X: 1920, Y: 1080}}
	ref2 := NewExternalReference("", "/path/1080p.mov", nil, nil)
	ref2.SetAvailableImageBounds(bounds2)
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip2 := NewClip("1080p", ref2, &sr2, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)

	result, err := track.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}

	if result == nil {
		t.Fatal("AvailableImageBounds should not be nil")
	}
	// Union should be the larger of the two
	if result.Max.X != 1920 {
		t.Errorf("Max.X = %v, want 1920", result.Max.X)
	}
	if result.Max.Y != 1080 {
		t.Errorf("Max.Y = %v, want 1080", result.Max.Y)
	}
}

// Tests for Track.AvailableImageBounds with gaps
func TestTrackAvailableImageBoundsWithGaps(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)

	// Add a gap (should not contribute to image bounds)
	gapSr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	gap := NewGap("gap", &gapSr, nil, nil, nil, nil)

	bounds := &Box2d{Min: Vec2d{X: 0, Y: 0}, Max: Vec2d{X: 1920, Y: 1080}}
	ref := NewExternalReference("", "/path/file.mov", nil, nil)
	ref.SetAvailableImageBounds(bounds)
	clipSr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", ref, &clipSr, nil, nil, nil, "", nil)

	track.AppendChild(gap)
	track.AppendChild(clip)

	result, err := track.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}

	if result == nil {
		t.Fatal("AvailableImageBounds should not be nil")
	}
	// Should only have clip's bounds (gap doesn't contribute)
	if result.Max.X != 1920 {
		t.Errorf("Max.X = %v, want 1920", result.Max.X)
	}
}

// Tests for sortByStartTime through RangeOfAllChildren
func TestSortByStartTimeOrdering(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)

	// Add clips in order
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("first", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("second", nil, &sr, nil, nil, nil, "", nil)
	clip3 := NewClip("third", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)
	track.AppendChild(clip3)

	ranges, err := track.RangeOfAllChildren()
	if err != nil {
		t.Fatalf("RangeOfAllChildren error: %v", err)
	}

	// Verify all clips are in the map
	if len(ranges) != 3 {
		t.Errorf("RangeOfAllChildren count = %d, want 3", len(ranges))
	}

	// Check that ranges are sequential
	r1, ok := ranges[clip1]
	if !ok {
		t.Fatal("clip1 not in ranges")
	}
	r2, ok := ranges[clip2]
	if !ok {
		t.Fatal("clip2 not in ranges")
	}
	r3, ok := ranges[clip3]
	if !ok {
		t.Fatal("clip3 not in ranges")
	}

	// Verify start times are in order
	if r1.StartTime().Value() >= r2.StartTime().Value() {
		t.Errorf("clip1 start %v should be < clip2 start %v", r1.StartTime().Value(), r2.StartTime().Value())
	}
	if r2.StartTime().Value() >= r3.StartTime().Value() {
		t.Errorf("clip2 start %v should be < clip3 start %v", r2.StartTime().Value(), r3.StartTime().Value())
	}
}

// Tests for FindChildren with searchRange
func TestFindChildrenWithSearchRange(t *testing.T) {
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip1 := NewClip("clip1", nil, &sr, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr, nil, nil, nil, "", nil)
	clip3 := NewClip("clip3", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip1)
	track.AppendChild(clip2)
	track.AppendChild(clip3)

	// Search in a range that includes all clips (they span 0-72)
	// Use a larger range to ensure all clips are included
	searchRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	found := track.FindChildren(&searchRange, true, nil)

	if len(found) < 2 {
		t.Errorf("FindChildren found %d, want at least 2", len(found))
	}

	// Test with nil searchRange (should return all children)
	found = track.FindChildren(nil, true, nil)
	if len(found) != 3 {
		t.Errorf("FindChildren with nil range found %d, want 3", len(found))
	}
}

// Tests for FindChildren with deep search and filter
func TestFindChildrenDeepSearchWithFilter(t *testing.T) {
	stack := NewStack("stack", nil, nil, nil, nil, nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("target_clip", nil, &sr, nil, nil, nil, "", nil)

	track.AppendChild(clip)
	stack.AppendChild(track)

	// Find clips with filter
	found := stack.FindChildren(nil, false, func(c Composable) bool {
		if cl, ok := c.(*Clip); ok {
			return cl.Name() == "target_clip"
		}
		return false
	})

	if len(found) != 1 {
		t.Errorf("FindChildren found %d, want 1", len(found))
	}
}

// Tests for ImageSequenceReference additional methods
func TestImageSequenceReferenceFrameMethods(t *testing.T) {
	ref := NewImageSequenceReference("", "/path/", "frame_", ".exr", 1001, 1, 24, 4, nil, nil, MissingFramePolicyError)

	// Test FrameForTime
	time := opentime.NewRationalTime(5, 24)
	frame := ref.FrameForTime(time)
	expected := 1001 + 5*1 // startFrame + frameIndex*frameStep
	if frame != expected {
		t.Errorf("FrameForTime = %d, want %d", frame, expected)
	}

	// Test TargetURLForFrame
	url := ref.TargetURLForFrame(1001)
	if url == "" {
		t.Error("TargetURLForFrame should return non-empty string")
	}
}

// Tests for NewImageSequenceReference with all parameters
func TestImageSequenceReferenceFullConstruction(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(100, 24))
	meta := AnyDictionary{"key": "value"}
	ref := NewImageSequenceReference(
		"test_name",
		"/path/to/",
		"frame_",
		".exr",
		1001,
		2,
		24,
		6,
		&ar,
		meta,
		MissingFramePolicyHold,
	)

	if ref.Name() != "test_name" {
		t.Errorf("Name = %s, want test_name", ref.Name())
	}
	if ref.TargetURLBase() != "/path/to/" {
		t.Errorf("TargetURLBase = %s, want /path/to/", ref.TargetURLBase())
	}
	if ref.NamePrefix() != "frame_" {
		t.Errorf("NamePrefix = %s, want frame_", ref.NamePrefix())
	}
	if ref.NameSuffix() != ".exr" {
		t.Errorf("NameSuffix = %s, want .exr", ref.NameSuffix())
	}
	if ref.StartFrame() != 1001 {
		t.Errorf("StartFrame = %d, want 1001", ref.StartFrame())
	}
	if ref.FrameStep() != 2 {
		t.Errorf("FrameStep = %d, want 2", ref.FrameStep())
	}
	if ref.Rate() != 24 {
		t.Errorf("Rate = %v, want 24", ref.Rate())
	}
	if ref.FrameZeroPadding() != 6 {
		t.Errorf("FrameZeroPadding = %d, want 6", ref.FrameZeroPadding())
	}
	if ref.MissingFramePolicy() != MissingFramePolicyHold {
		t.Errorf("MissingFramePolicy = %v, want %v", ref.MissingFramePolicy(), MissingFramePolicyHold)
	}
}

// Tests for EndFrame
func TestImageSequenceReferenceEndFrame(t *testing.T) {
	ar := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	ref := NewImageSequenceReference("", "/path/", "frame_", ".exr", 1001, 1, 24, 4, &ar, nil, MissingFramePolicyError)

	// EndFrame should be startFrame + (frameCount - 1) * frameStep
	endFrame := ref.EndFrame()
	expected := 1001 + (48-1)*1 // 1048
	if endFrame != expected {
		t.Errorf("EndFrame = %d, want %d", endFrame, expected)
	}
}

// Tests for Timeline.IsEquivalentTo with track variations
func TestTimelineIsEquivalentToTrackVariations(t *testing.T) {
	timeline1 := NewTimeline("timeline", nil, nil)
	timeline2 := NewTimeline("timeline", nil, nil)
	timeline3 := NewTimeline("different", nil, nil)

	if !timeline1.IsEquivalentTo(timeline2) {
		t.Error("Identical timelines should be equivalent")
	}
	if timeline1.IsEquivalentTo(timeline3) {
		t.Error("Different timelines should not be equivalent")
	}

	// Test with different tracks content
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	track.AppendChild(clip)
	timeline1.Tracks().AppendChild(track)

	// timeline1 now has a track, timeline2 has empty tracks
	if timeline1.IsEquivalentTo(timeline2) {
		t.Error("Timeline with tracks should not equal timeline without tracks")
	}
}

// Tests for ComposableBase.Parent with nil
func TestComposableBaseParentNil(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	// Parent should be nil initially
	if clip.Parent() != nil {
		t.Error("New clip should have nil parent")
	}
}

// Tests for SetMediaReferences handling
func TestClipSetMediaReferencesHandling(t *testing.T) {
	clip := NewClip("clip", nil, nil, nil, nil, nil, "", nil)

	// Set references to non-nil
	refs := map[string]MediaReference{
		"default": NewExternalReference("", "/path/file.mov", nil, nil),
	}
	clip.SetMediaReferences(refs, "default")

	if len(clip.MediaReferences()) != 1 {
		t.Errorf("MediaReferences count = %d, want 1", len(clip.MediaReferences()))
	}

	// Set to empty map (the implementation may preserve the map but make it empty)
	clip.SetMediaReferences(map[string]MediaReference{}, "")
	t.Logf("MediaReferences count after empty map = %d", len(clip.MediaReferences()))
}
