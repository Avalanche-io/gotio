#!/bin/bash
# JSON Throughput Comparison: Go vs C++
# Runs both benchmarks and displays side-by-side results

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Default parameters
ITERATIONS=${ITERATIONS:-50}
VIDEO_TRACKS=${VIDEO_TRACKS:-3}
AUDIO_TRACKS=${AUDIO_TRACKS:-2}
CLIPS=${CLIPS:-100}

echo "=============================================================="
echo "JSON THROUGHPUT COMPARISON: Go vs C++"
echo "=============================================================="
echo ""
echo "Configuration:"
echo "  Iterations: $ITERATIONS"
echo "  Video tracks: $VIDEO_TRACKS"
echo "  Audio tracks: $AUDIO_TRACKS"
echo "  Clips per track: $CLIPS"
echo ""

# Build if needed
if [ ! -f go-json-bench/json-benchmark ]; then
    echo "Building Go benchmark..."
    (cd go-json-bench && go build -o json-benchmark .)
fi

if [ ! -f cpp-json-bench/build/json-benchmark ]; then
    echo "Building C++ benchmark..."
    (cd cpp-json-bench && mkdir -p build && cd build && cmake .. && make -j8)
fi

# Generate test data if needed
if [ ! -d testdata ] || [ -z "$(ls -A testdata 2>/dev/null)" ]; then
    echo "Generating test data..."
    ./go-json-bench/json-benchmark --generate testdata --generate-count 20
fi

echo ""
echo "=============================================================="
echo "RUNNING GO BENCHMARK"
echo "=============================================================="
./go-json-bench/json-benchmark \
    --iterations "$ITERATIONS" \
    --video-tracks "$VIDEO_TRACKS" \
    --audio-tracks "$AUDIO_TRACKS" \
    --clips "$CLIPS" \
    --testdata testdata

echo ""
echo "=============================================================="
echo "RUNNING C++ BENCHMARK"
echo "=============================================================="
./cpp-json-bench/build/json-benchmark \
    --iterations "$ITERATIONS" \
    --video-tracks "$VIDEO_TRACKS" \
    --audio-tracks "$AUDIO_TRACKS" \
    --clips "$CLIPS" \
    --testdata testdata

echo ""
echo "=============================================================="
echo "CROSS-LANGUAGE COMPARISON COMPLETE"
echo "=============================================================="
