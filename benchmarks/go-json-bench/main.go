// JSON Throughput Benchmark - Go Implementation
// Compares stdlib, jsoniter, go-json, and sonic JSON libraries

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	gojson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

// Benchmark result
type BenchmarkResult struct {
	Library       string
	Operation     string
	Iterations    int
	TotalBytes    int64
	TotalDuration time.Duration
	ThroughputMBs float64
	AvgLatencyUs  float64
}

// JSON libraries to test
type JSONLibrary struct {
	Name      string
	Marshal   func(v any) ([]byte, error)
	Unmarshal func(data []byte, v any) error
}

var libraries = []JSONLibrary{
	{
		Name:      "encoding/json",
		Marshal:   json.Marshal,
		Unmarshal: json.Unmarshal,
	},
	{
		Name:      "jsoniter",
		Marshal:   jsoniter.Marshal,
		Unmarshal: jsoniter.Unmarshal,
	},
	{
		Name:      "go-json",
		Marshal:   gojson.Marshal,
		Unmarshal: gojson.Unmarshal,
	},
}

// OTIO-like structures for realistic testing
type RationalTime struct {
	Schema string  `json:"OTIO_SCHEMA"`
	Value  float64 `json:"value"`
	Rate   float64 `json:"rate"`
}

type TimeRange struct {
	Schema    string       `json:"OTIO_SCHEMA"`
	StartTime RationalTime `json:"start_time"`
	Duration  RationalTime `json:"duration"`
}

type Clip struct {
	Schema           string            `json:"OTIO_SCHEMA"`
	Name             string            `json:"name"`
	SourceRange      *TimeRange        `json:"source_range"`
	MediaReference   *MediaReference   `json:"media_reference"`
	Metadata         map[string]any    `json:"metadata"`
	ActiveMediaRef   string            `json:"active_media_reference_key"`
	Markers          []Marker          `json:"markers"`
	Effects          []Effect          `json:"effects"`
	Enabled          bool              `json:"enabled"`
}

type MediaReference struct {
	Schema        string         `json:"OTIO_SCHEMA"`
	Name          string         `json:"name"`
	AvailableRange *TimeRange    `json:"available_range"`
	TargetURL     string         `json:"target_url"`
	Metadata      map[string]any `json:"metadata"`
}

type Marker struct {
	Schema      string         `json:"OTIO_SCHEMA"`
	Name        string         `json:"name"`
	MarkedRange TimeRange      `json:"marked_range"`
	Color       string         `json:"color"`
	Metadata    map[string]any `json:"metadata"`
}

type Effect struct {
	Schema   string         `json:"OTIO_SCHEMA"`
	Name     string         `json:"name"`
	Metadata map[string]any `json:"metadata"`
}

type Track struct {
	Schema   string         `json:"OTIO_SCHEMA"`
	Name     string         `json:"name"`
	Kind     string         `json:"kind"`
	Children []Clip         `json:"children"`
	Metadata map[string]any `json:"metadata"`
}

type Stack struct {
	Schema   string         `json:"OTIO_SCHEMA"`
	Name     string         `json:"name"`
	Children []Track        `json:"children"`
	Metadata map[string]any `json:"metadata"`
}

type Timeline struct {
	Schema      string         `json:"OTIO_SCHEMA"`
	Name        string         `json:"name"`
	GlobalStart RationalTime   `json:"global_start_time"`
	Tracks      Stack          `json:"tracks"`
	Metadata    map[string]any `json:"metadata"`
}

func generateClip(index int) Clip {
	sr := TimeRange{
		Schema: "TimeRange.1",
		StartTime: RationalTime{
			Schema: "RationalTime.1",
			Value:  float64(index * 24),
			Rate:   24.0,
		},
		Duration: RationalTime{
			Schema: "RationalTime.1",
			Value:  48.0,
			Rate:   24.0,
		},
	}

	mr := MediaReference{
		Schema: "ExternalReference.1",
		Name:   fmt.Sprintf("media_%d", index),
		AvailableRange: &TimeRange{
			Schema: "TimeRange.1",
			StartTime: RationalTime{
				Schema: "RationalTime.1",
				Value:  0,
				Rate:   24.0,
			},
			Duration: RationalTime{
				Schema: "RationalTime.1",
				Value:  1000,
				Rate:   24.0,
			},
		},
		TargetURL: fmt.Sprintf("file:///media/project/footage/clip_%05d.mov", index),
		Metadata: map[string]any{
			"codec":      "ProRes422HQ",
			"resolution": "1920x1080",
			"colorspace": "Rec709",
		},
	}

	return Clip{
		Schema:         "Clip.2",
		Name:           fmt.Sprintf("Shot_%04d", index),
		SourceRange:    &sr,
		MediaReference: &mr,
		Metadata: map[string]any{
			"shot_type":  "wide",
			"scene":      fmt.Sprintf("Scene_%d", index/10),
			"take":       index % 5,
			"notes":      "This is a sample note for the clip with some additional text to make it more realistic.",
			"color_tag":  "green",
			"approved":   true,
			"frame_rate": 24.0,
		},
		ActiveMediaRef: "DEFAULT_MEDIA",
		Markers:        []Marker{},
		Effects:        []Effect{},
		Enabled:        true,
	}
}

func generateTrack(name string, kind string, clipCount int) Track {
	clips := make([]Clip, clipCount)
	for i := 0; i < clipCount; i++ {
		clips[i] = generateClip(i)
	}
	return Track{
		Schema:   "Track.1",
		Name:     name,
		Kind:     kind,
		Children: clips,
		Metadata: map[string]any{
			"track_index": 0,
			"locked":      false,
			"muted":       false,
		},
	}
}

func generateTimeline(name string, videoTracks, audioTracks, clipsPerTrack int) Timeline {
	tracks := make([]Track, videoTracks+audioTracks)
	for i := 0; i < videoTracks; i++ {
		tracks[i] = generateTrack(fmt.Sprintf("V%d", i+1), "Video", clipsPerTrack)
	}
	for i := 0; i < audioTracks; i++ {
		tracks[videoTracks+i] = generateTrack(fmt.Sprintf("A%d", i+1), "Audio", clipsPerTrack)
	}

	return Timeline{
		Schema: "Timeline.1",
		Name:   name,
		GlobalStart: RationalTime{
			Schema: "RationalTime.1",
			Value:  86400.0,
			Rate:   24.0,
		},
		Tracks: Stack{
			Schema:   "Stack.1",
			Name:     "tracks",
			Children: tracks,
			Metadata: map[string]any{},
		},
		Metadata: map[string]any{
			"project":      "Benchmark Project",
			"created_by":   "json-benchmark",
			"created_date": time.Now().Format(time.RFC3339),
		},
	}
}

func runMarshalBenchmark(lib JSONLibrary, data any, iterations int) BenchmarkResult {
	// Warmup
	for i := 0; i < 10; i++ {
		lib.Marshal(data)
	}

	runtime.GC()

	var totalBytes int64
	start := time.Now()

	for i := 0; i < iterations; i++ {
		bytes, err := lib.Marshal(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Marshal error: %v\n", err)
			continue
		}
		totalBytes += int64(len(bytes))
	}

	elapsed := time.Since(start)
	throughput := float64(totalBytes) / elapsed.Seconds() / (1024 * 1024)
	avgLatency := float64(elapsed.Microseconds()) / float64(iterations)

	return BenchmarkResult{
		Library:       lib.Name,
		Operation:     "Marshal",
		Iterations:    iterations,
		TotalBytes:    totalBytes,
		TotalDuration: elapsed,
		ThroughputMBs: throughput,
		AvgLatencyUs:  avgLatency,
	}
}

func runUnmarshalBenchmark(lib JSONLibrary, jsonData []byte, iterations int) BenchmarkResult {
	// Warmup
	for i := 0; i < 10; i++ {
		var t Timeline
		lib.Unmarshal(jsonData, &t)
	}

	runtime.GC()

	totalBytes := int64(len(jsonData)) * int64(iterations)
	start := time.Now()

	for i := 0; i < iterations; i++ {
		var t Timeline
		if err := lib.Unmarshal(jsonData, &t); err != nil {
			fmt.Fprintf(os.Stderr, "Unmarshal error: %v\n", err)
			continue
		}
	}

	elapsed := time.Since(start)
	throughput := float64(totalBytes) / elapsed.Seconds() / (1024 * 1024)
	avgLatency := float64(elapsed.Microseconds()) / float64(iterations)

	return BenchmarkResult{
		Library:       lib.Name,
		Operation:     "Unmarshal",
		Iterations:    iterations,
		TotalBytes:    totalBytes,
		TotalDuration: elapsed,
		ThroughputMBs: throughput,
		AvgLatencyUs:  avgLatency,
	}
}

func loadTestDataFiles(dir string) ([][]byte, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}

	var data [][]byte
	for _, f := range files {
		bytes, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		data = append(data, bytes)
	}
	return data, nil
}

func runFileBenchmark(lib JSONLibrary, files [][]byte, iterations int) (BenchmarkResult, BenchmarkResult) {
	// Warmup
	for i := 0; i < 5; i++ {
		for _, f := range files {
			var t Timeline
			lib.Unmarshal(f, &t)
		}
	}

	runtime.GC()

	// Unmarshal benchmark
	var totalBytes int64
	for _, f := range files {
		totalBytes += int64(len(f))
	}
	totalBytes *= int64(iterations)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		for _, f := range files {
			var t Timeline
			lib.Unmarshal(f, &t)
		}
	}
	unmarshalElapsed := time.Since(start)

	unmarshalResult := BenchmarkResult{
		Library:       lib.Name,
		Operation:     "Unmarshal (files)",
		Iterations:    iterations * len(files),
		TotalBytes:    totalBytes,
		TotalDuration: unmarshalElapsed,
		ThroughputMBs: float64(totalBytes) / unmarshalElapsed.Seconds() / (1024 * 1024),
		AvgLatencyUs:  float64(unmarshalElapsed.Microseconds()) / float64(iterations*len(files)),
	}

	// Marshal benchmark - parse files first
	var timelines []Timeline
	for _, f := range files {
		var t Timeline
		lib.Unmarshal(f, &t)
		timelines = append(timelines, t)
	}

	runtime.GC()

	totalBytes = 0
	start = time.Now()
	for i := 0; i < iterations; i++ {
		for _, t := range timelines {
			bytes, _ := lib.Marshal(t)
			totalBytes += int64(len(bytes))
		}
	}
	marshalElapsed := time.Since(start)

	marshalResult := BenchmarkResult{
		Library:       lib.Name,
		Operation:     "Marshal (files)",
		Iterations:    iterations * len(timelines),
		TotalBytes:    totalBytes,
		TotalDuration: marshalElapsed,
		ThroughputMBs: float64(totalBytes) / marshalElapsed.Seconds() / (1024 * 1024),
		AvgLatencyUs:  float64(marshalElapsed.Microseconds()) / float64(iterations*len(timelines)),
	}

	return marshalResult, unmarshalResult
}

func printResults(results []BenchmarkResult) {
	// Sort by operation then throughput
	sort.Slice(results, func(i, j int) bool {
		if results[i].Operation != results[j].Operation {
			return results[i].Operation < results[j].Operation
		}
		return results[i].ThroughputMBs > results[j].ThroughputMBs
	})

	separator := strings.Repeat("=", 80)
	line := strings.Repeat("-", 80)

	fmt.Println("\n" + separator)
	fmt.Println("BENCHMARK RESULTS")
	fmt.Println(separator)
	fmt.Printf("%-20s %-20s %12s %12s %12s\n", "Library", "Operation", "Throughput", "Avg Latency", "Total MB")
	fmt.Println(line)

	currentOp := ""
	for _, r := range results {
		if r.Operation != currentOp {
			if currentOp != "" {
				fmt.Println(line)
			}
			currentOp = r.Operation
		}
		fmt.Printf("%-20s %-20s %9.2f MB/s %9.2f us %9.2f MB\n",
			r.Library,
			r.Operation,
			r.ThroughputMBs,
			r.AvgLatencyUs,
			float64(r.TotalBytes)/(1024*1024),
		)
	}
	fmt.Println(separator)
}

func main() {
	var (
		videoTracks   = flag.Int("video-tracks", 3, "Number of video tracks")
		audioTracks   = flag.Int("audio-tracks", 2, "Number of audio tracks")
		clipsPerTrack = flag.Int("clips", 100, "Clips per track")
		iterations    = flag.Int("iterations", 100, "Number of iterations")
		testDataDir   = flag.String("testdata", "", "Directory with .json test files")
		generateOnly  = flag.String("generate", "", "Generate test data to directory and exit")
		generateCount = flag.Int("generate-count", 10, "Number of test files to generate")
	)
	flag.Parse()

	// Generate test data mode
	if *generateOnly != "" {
		fmt.Printf("Generating %d test files to %s\n", *generateCount, *generateOnly)
		os.MkdirAll(*generateOnly, 0755)

		configs := []struct {
			video, audio, clips int
			name                string
		}{
			{1, 1, 10, "small"},
			{2, 2, 50, "medium"},
			{3, 2, 100, "standard"},
			{5, 4, 200, "large"},
			{10, 8, 500, "xlarge"},
		}

		for i := 0; i < *generateCount; i++ {
			cfg := configs[i%len(configs)]
			timeline := generateTimeline(
				fmt.Sprintf("Timeline_%s_%d", cfg.name, i),
				cfg.video, cfg.audio, cfg.clips,
			)
			data, _ := json.MarshalIndent(timeline, "", "  ")
			filename := filepath.Join(*generateOnly, fmt.Sprintf("timeline_%s_%03d.json", cfg.name, i))
			os.WriteFile(filename, data, 0644)
			fmt.Printf("  Generated %s (%d bytes)\n", filename, len(data))
		}
		return
	}

	fmt.Println("Go JSON Throughput Benchmark")
	fmt.Println("============================")
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("CPUs: %d\n", runtime.NumCPU())

	var results []BenchmarkResult

	// Test with generated data
	fmt.Printf("\nGenerating timeline: %d video + %d audio tracks, %d clips each\n",
		*videoTracks, *audioTracks, *clipsPerTrack)

	timeline := generateTimeline("Benchmark Timeline", *videoTracks, *audioTracks, *clipsPerTrack)

	// Get baseline JSON size
	baselineJSON, _ := json.Marshal(timeline)
	fmt.Printf("Timeline JSON size: %.2f MB\n", float64(len(baselineJSON))/(1024*1024))
	fmt.Printf("Running %d iterations per library\n", *iterations)

	for _, lib := range libraries {
		fmt.Printf("\nBenchmarking %s...\n", lib.Name)

		// Marshal
		result := runMarshalBenchmark(lib, timeline, *iterations)
		results = append(results, result)
		fmt.Printf("  Marshal: %.2f MB/s\n", result.ThroughputMBs)

		// Unmarshal
		result = runUnmarshalBenchmark(lib, baselineJSON, *iterations)
		results = append(results, result)
		fmt.Printf("  Unmarshal: %.2f MB/s\n", result.ThroughputMBs)
	}

	// Test with file data if provided
	if *testDataDir != "" {
		fmt.Printf("\nLoading test files from %s\n", *testDataDir)
		files, err := loadTestDataFiles(*testDataDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading test files: %v\n", err)
		} else if len(files) > 0 {
			var totalSize int64
			for _, f := range files {
				totalSize += int64(len(f))
			}
			fmt.Printf("Loaded %d files, total %.2f MB\n", len(files), float64(totalSize)/(1024*1024))

			for _, lib := range libraries {
				fmt.Printf("\nBenchmarking %s with files...\n", lib.Name)
				marshal, unmarshal := runFileBenchmark(lib, files, *iterations/10)
				results = append(results, marshal, unmarshal)
				fmt.Printf("  Marshal: %.2f MB/s, Unmarshal: %.2f MB/s\n",
					marshal.ThroughputMBs, unmarshal.ThroughputMBs)
			}
		}
	}

	printResults(results)

	// Output summary for comparison
	separator := strings.Repeat("=", 80)
	fmt.Println("\n" + separator)
	fmt.Println("SUMMARY FOR CROSS-LANGUAGE COMPARISON")
	fmt.Println(separator)
	fmt.Printf("Data size: %.2f MB\n", float64(len(baselineJSON))/(1024*1024))
	fmt.Printf("Iterations: %d\n", *iterations)

	// Find best results
	var bestMarshal, bestUnmarshal BenchmarkResult
	for _, r := range results {
		if r.Operation == "Marshal" && r.ThroughputMBs > bestMarshal.ThroughputMBs {
			bestMarshal = r
		}
		if r.Operation == "Unmarshal" && r.ThroughputMBs > bestUnmarshal.ThroughputMBs {
			bestUnmarshal = r
		}
	}

	fmt.Printf("\nBest Marshal: %s at %.2f MB/s (%.2f us/op)\n",
		bestMarshal.Library, bestMarshal.ThroughputMBs, bestMarshal.AvgLatencyUs)
	fmt.Printf("Best Unmarshal: %s at %.2f MB/s (%.2f us/op)\n",
		bestUnmarshal.Library, bestUnmarshal.ThroughputMBs, bestUnmarshal.AvgLatencyUs)
}
