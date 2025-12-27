// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// otiogen generates high-performance JSON encoders and decoders for OTIO types.
//
// Usage:
//
//	go run ./cmd/otiogen
//
// This generates:
//   - internal/jsonenc/gen_opentime.go  - opentime type encoders
//   - internal/jsonenc/gen_otio.go      - OTIO leaf type encoders
//   - internal/jsondec/gen_opentime.go  - opentime type decoders
//   - internal/jsondec/gen_otio.go      - OTIO leaf type decoders
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	outputDir := flag.String("output", ".", "Output directory (project root)")
	flag.Parse()

	if err := run(*outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "otiogen: %v\n", err)
		os.Exit(1)
	}
}

func run(outputDir string) error {
	gen, err := NewGenerator()
	if err != nil {
		return err
	}

	// Ensure output directories exist
	encDir := filepath.Join(outputDir, "internal", "jsonenc")
	decDir := filepath.Join(outputDir, "internal", "jsondec")

	if err := os.MkdirAll(encDir, 0755); err != nil {
		return fmt.Errorf("create encoder dir: %w", err)
	}
	if err := os.MkdirAll(decDir, 0755); err != nil {
		return fmt.Errorf("create decoder dir: %w", err)
	}

	// Generate opentime encoders
	fmt.Println("Generating opentime encoders...")
	openTimeEnc, err := gen.GenerateOpenTimeEncoders()
	if err != nil {
		return fmt.Errorf("generate opentime encoders: %w", err)
	}
	if err := os.WriteFile(filepath.Join(encDir, "gen_opentime.go"), openTimeEnc, 0644); err != nil {
		return fmt.Errorf("write opentime encoders: %w", err)
	}

	// Generate opentime decoders
	fmt.Println("Generating opentime decoders...")
	openTimeDec, err := gen.GenerateOpenTimeDecoders()
	if err != nil {
		return fmt.Errorf("generate opentime decoders: %w", err)
	}
	if err := os.WriteFile(filepath.Join(decDir, "gen_opentime.go"), openTimeDec, 0644); err != nil {
		return fmt.Errorf("write opentime decoders: %w", err)
	}

	// Generate OTIO leaf encoders
	fmt.Println("Generating OTIO leaf encoders...")
	otioEnc, err := gen.GenerateOTIOLeafEncoders()
	if err != nil {
		return fmt.Errorf("generate OTIO encoders: %w", err)
	}
	if err := os.WriteFile(filepath.Join(encDir, "gen_otio.go"), otioEnc, 0644); err != nil {
		return fmt.Errorf("write OTIO encoders: %w", err)
	}

	// Generate OTIO leaf decoders
	fmt.Println("Generating OTIO leaf decoders...")
	otioDec, err := gen.GenerateOTIOLeafDecoders()
	if err != nil {
		return fmt.Errorf("generate OTIO decoders: %w", err)
	}
	if err := os.WriteFile(filepath.Join(decDir, "gen_otio.go"), otioDec, 0644); err != nil {
		return fmt.Errorf("write OTIO decoders: %w", err)
	}

	fmt.Println("Done!")
	return nil
}
