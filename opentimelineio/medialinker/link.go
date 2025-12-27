// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package medialinker

import (
	"github.com/mrjoshuak/gotio/opentimelineio"
)

// LinkConfig holds configuration for the linking operation.
type LinkConfig struct {
	// ContinueOnError continues linking even if some clips fail.
	ContinueOnError bool
	// Args are passed to the linker for each clip.
	Args map[string]any
}

// LinkOption is a functional option for LinkMedia.
type LinkOption func(*LinkConfig)

// WithContinueOnError sets whether to continue on errors.
func WithContinueOnError(continueOnError bool) LinkOption {
	return func(c *LinkConfig) {
		c.ContinueOnError = continueOnError
	}
}

// WithArgs sets the arguments to pass to the linker.
func WithArgs(args map[string]any) LinkOption {
	return func(c *LinkConfig) {
		c.Args = args
	}
}

// LinkMedia applies a media linker to all clips in a timeline.
// Uses the named linker from the global registry.
func LinkMedia(
	timeline *opentimelineio.Timeline,
	linkerName string,
	opts ...LinkOption,
) error {
	linker, err := Get(linkerName)
	if err != nil {
		return err
	}
	return LinkMediaWithLinker(timeline, linker, opts...)
}

// LinkMediaDefault applies the default media linker to all clips.
func LinkMediaDefault(
	timeline *opentimelineio.Timeline,
	opts ...LinkOption,
) error {
	linker, err := Default()
	if err != nil {
		return err
	}
	return LinkMediaWithLinker(timeline, linker, opts...)
}

// LinkMediaWithLinker applies a specific linker to all clips in a timeline.
func LinkMediaWithLinker(
	timeline *opentimelineio.Timeline,
	linker MediaLinker,
	opts ...LinkOption,
) error {
	// Apply options
	config := &LinkConfig{
		ContinueOnError: false,
		Args:            make(map[string]any),
	}
	for _, opt := range opts {
		opt(config)
	}

	// Find all clips
	clips := timeline.FindClips(nil, false)

	var lastError error

	for _, clip := range clips {
		newRef, err := linker.LinkMediaReference(clip, config.Args)
		if err != nil {
			lastError = &LinkerError{
				LinkerName: linker.Name(),
				ClipName:   clip.Name(),
				Message:    "linking failed",
				Cause:      err,
			}
			if !config.ContinueOnError {
				return lastError
			}
			continue
		}

		// Only update if a new reference was returned
		if newRef != nil {
			clip.SetMediaReference(newRef)
		}
	}

	return lastError
}

// LinkClip applies a media linker to a single clip.
func LinkClip(
	clip *opentimelineio.Clip,
	linker MediaLinker,
	args map[string]any,
) error {
	if args == nil {
		args = make(map[string]any)
	}

	newRef, err := linker.LinkMediaReference(clip, args)
	if err != nil {
		return &LinkerError{
			LinkerName: linker.Name(),
			ClipName:   clip.Name(),
			Message:    "linking failed",
			Cause:      err,
		}
	}

	if newRef != nil {
		clip.SetMediaReference(newRef)
	}

	return nil
}
