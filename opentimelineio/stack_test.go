// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package opentimelineio

import (
	"encoding/json"
	"testing"

	"github.com/Avalanche-io/gotio/opentime"
)

func TestStackCompositionKind(t *testing.T) {
	stack := NewStack("test", nil, nil, nil, nil, nil)

	if stack.CompositionKind() != "Stack" {
		t.Errorf("CompositionKind = %s, want Stack", stack.CompositionKind())
	}
}

func TestStackRangeOfChildAtIndex(t *testing.T) {
	stack := NewStack("test", nil, nil, nil, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	stack.AppendChild(clip1)
	stack.AppendChild(clip2)

	// In a stack, all children start at time 0
	r1, err := stack.RangeOfChildAtIndex(0)
	if err != nil {
		t.Fatalf("RangeOfChildAtIndex(0) error: %v", err)
	}
	if r1.StartTime().Value() != 0 {
		t.Errorf("Child 0 start time = %v, want 0", r1.StartTime().Value())
	}
	if r1.Duration().Value() != 24 {
		t.Errorf("Child 0 duration = %v, want 24", r1.Duration().Value())
	}

	r2, err := stack.RangeOfChildAtIndex(1)
	if err != nil {
		t.Fatalf("RangeOfChildAtIndex(1) error: %v", err)
	}
	if r2.StartTime().Value() != 0 {
		t.Errorf("Child 1 start time = %v, want 0", r2.StartTime().Value())
	}
	if r2.Duration().Value() != 48 {
		t.Errorf("Child 1 duration = %v, want 48", r2.Duration().Value())
	}

	// Test invalid index
	_, err = stack.RangeOfChildAtIndex(-1)
	if err == nil {
		t.Error("RangeOfChildAtIndex(-1) should error")
	}
	_, err = stack.RangeOfChildAtIndex(10)
	if err == nil {
		t.Error("RangeOfChildAtIndex(10) should error")
	}
}

func TestStackAvailableRange(t *testing.T) {
	stack := NewStack("test", nil, nil, nil, nil, nil)

	// Empty stack
	ar, err := stack.AvailableRange()
	if err != nil {
		t.Fatalf("AvailableRange error: %v", err)
	}
	if ar.Duration().Rate() != 0 {
		t.Errorf("Empty stack should have zero duration")
	}

	// Add children with different durations
	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(72, 24))

	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	stack.AppendChild(clip1)
	stack.AppendChild(clip2)

	// Stack duration is the max of all children
	ar, err = stack.AvailableRange()
	if err != nil {
		t.Fatalf("AvailableRange error: %v", err)
	}
	if ar.Duration().Value() != 72 {
		t.Errorf("AvailableRange duration = %v, want 72", ar.Duration().Value())
	}
}

func TestStackDuration(t *testing.T) {
	stack := NewStack("test", nil, nil, nil, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(96, 24))

	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	stack.AppendChild(clip1)
	stack.AppendChild(clip2)

	dur, err := stack.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	// Duration is max duration (96 frames = 4 seconds)
	if dur.ToSeconds() != 4.0 {
		t.Errorf("Duration = %v seconds, want 4.0", dur.ToSeconds())
	}

	// Test with source range
	srStack := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	stack2 := NewStack("test2", &srStack, nil, nil, nil, nil)
	stack2.AppendChild(clip1)
	stack2.AppendChild(clip2)

	dur, err = stack2.Duration()
	if err != nil {
		t.Fatalf("Duration error: %v", err)
	}
	// Source range overrides computed duration
	if dur.Value() != 24 {
		t.Errorf("Duration with source range = %v, want 24", dur.Value())
	}
}

func TestStackChildAtTime(t *testing.T) {
	stack := NewStack("test", nil, nil, nil, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	stack.AppendChild(clip1)
	stack.AppendChild(clip2)

	// At time 0, both clips exist, should return topmost (clip2)
	child, err := stack.ChildAtTime(opentime.NewRationalTime(0, 24), true)
	if err != nil {
		t.Fatalf("ChildAtTime error: %v", err)
	}
	if child.(*Clip).Name() != "clip2" {
		t.Errorf("ChildAtTime(0) = %s, want clip2", child.(*Clip).Name())
	}

	// At time 30 frames (1.25 seconds), only clip2 exists
	child, err = stack.ChildAtTime(opentime.NewRationalTime(30, 24), true)
	if err != nil {
		t.Fatalf("ChildAtTime error: %v", err)
	}
	if child.(*Clip).Name() != "clip2" {
		t.Errorf("ChildAtTime(30) = %s, want clip2", child.(*Clip).Name())
	}

	// Outside range
	child, err = stack.ChildAtTime(opentime.NewRationalTime(100, 24), true)
	if err != nil {
		t.Fatalf("ChildAtTime error: %v", err)
	}
	if child != nil {
		t.Error("ChildAtTime outside range should return nil")
	}
}

func TestStackChildrenInRange(t *testing.T) {
	stack := NewStack("test", nil, nil, nil, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	stack.AppendChild(clip1)
	stack.AppendChild(clip2)

	// Range 0-12 frames - both clips overlap
	searchRange := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(12, 24))
	children, err := stack.ChildrenInRange(searchRange)
	if err != nil {
		t.Fatalf("ChildrenInRange error: %v", err)
	}
	if len(children) != 2 {
		t.Errorf("ChildrenInRange(0-12) = %d children, want 2", len(children))
	}

	// Range 24-48 frames - only clip2
	searchRange = opentime.NewTimeRange(opentime.NewRationalTime(24, 24), opentime.NewRationalTime(24, 24))
	children, err = stack.ChildrenInRange(searchRange)
	if err != nil {
		t.Fatalf("ChildrenInRange error: %v", err)
	}
	if len(children) != 1 {
		t.Errorf("ChildrenInRange(24-48) = %d children, want 1", len(children))
	}
}

func TestStackRangeOfAllChildren(t *testing.T) {
	stack := NewStack("test", nil, nil, nil, nil, nil)

	sr1 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	sr2 := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))

	clip1 := NewClip("clip1", nil, &sr1, nil, nil, nil, "", nil)
	clip2 := NewClip("clip2", nil, &sr2, nil, nil, nil, "", nil)

	stack.AppendChild(clip1)
	stack.AppendChild(clip2)

	ranges, err := stack.RangeOfAllChildren()
	if err != nil {
		t.Fatalf("RangeOfAllChildren error: %v", err)
	}
	if len(ranges) != 2 {
		t.Errorf("RangeOfAllChildren = %d entries, want 2", len(ranges))
	}

	// All children should start at 0
	for _, r := range ranges {
		if r.StartTime().Value() != 0 {
			t.Errorf("Child start time = %v, want 0", r.StartTime().Value())
		}
	}
}

func TestStackClone(t *testing.T) {
	stack := NewStack("test", nil, AnyDictionary{"key": "value"}, nil, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	stack.AppendChild(clip)

	clone := stack.Clone().(*Stack)

	if clone.Name() != "test" {
		t.Errorf("Clone name = %s, want test", clone.Name())
	}
	if len(clone.Children()) != 1 {
		t.Errorf("Clone children count = %d, want 1", len(clone.Children()))
	}

	// Verify deep copy
	clone.SetName("modified")
	if stack.Name() == "modified" {
		t.Error("Modifying clone affected original")
	}
}

func TestStackIsEquivalentTo(t *testing.T) {
	s1 := NewStack("test", nil, nil, nil, nil, nil)
	s2 := NewStack("test", nil, nil, nil, nil, nil)
	s3 := NewStack("different", nil, nil, nil, nil, nil)

	if !s1.IsEquivalentTo(s2) {
		t.Error("Identical stacks should be equivalent")
	}
	if s1.IsEquivalentTo(s3) {
		t.Error("Different stacks should not be equivalent")
	}

	// Test with different children count
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	s1.AppendChild(clip)
	if s1.IsEquivalentTo(s2) {
		t.Error("Stacks with different children should not be equivalent")
	}

	// Test with non-Stack
	track := NewTrack("track", nil, TrackKindVideo, nil, nil)
	if s1.IsEquivalentTo(track) {
		t.Error("Stack should not be equivalent to Track")
	}
}

func TestStackJSON(t *testing.T) {
	stack := NewStack("test_stack", nil, AnyDictionary{"note": "test"}, nil, nil, nil)

	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(48, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	stack.AppendChild(clip)

	data, err := json.Marshal(stack)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	stack2 := &Stack{}
	if err := json.Unmarshal(data, stack2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if stack2.Name() != "test_stack" {
		t.Errorf("Name mismatch: got %s", stack2.Name())
	}
	if len(stack2.Children()) != 1 {
		t.Errorf("Children count mismatch: got %d", len(stack2.Children()))
	}
}

func TestStackAvailableImageBounds(t *testing.T) {
	stack := NewStack("test", nil, nil, nil, nil, nil)

	// Test with empty stack
	bounds, err := stack.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}
	if bounds != nil {
		t.Error("Empty stack should have nil bounds")
	}

	// Add a clip without image bounds
	sr := opentime.NewTimeRange(opentime.NewRationalTime(0, 24), opentime.NewRationalTime(24, 24))
	clip := NewClip("clip", nil, &sr, nil, nil, nil, "", nil)
	stack.AppendChild(clip)

	bounds, err = stack.AvailableImageBounds()
	if err != nil {
		t.Fatalf("AvailableImageBounds error: %v", err)
	}
	// Clip without media reference has no bounds
}
