# JSON Throughput Benchmark: Go vs C++

This benchmark suite compares JSON serialization/deserialization performance between Go and C++ implementations, using OTIO-like timeline data structures.

## Results Summary

**Test Environment:** Apple M3 Max, macOS, 16 cores

### Generated Data (0.44 MB timeline)

| Library | Language | Marshal/Stringify | Parse/Unmarshal |
|---------|----------|-------------------|-----------------|
| **go-json** | Go | **601 MB/s** | 446 MB/s |
| RapidJSON | C++ | 548 MB/s | **539 MB/s** |
| encoding/json | Go | 447 MB/s | 125 MB/s |
| jsoniter | Go | 395 MB/s | 296 MB/s |
| nlohmann/json | C++ | 329 MB/s | 136 MB/s |

### File Data (78 MB across 20 files)

| Library | Language | Marshal/Stringify | Parse/Unmarshal |
|---------|----------|-------------------|-----------------|
| **go-json** | Go | **616 MB/s** | 773 MB/s |
| RapidJSON | C++ | N/A | **931 MB/s** |
| jsoniter | Go | 398 MB/s | 505 MB/s |
| encoding/json | Go | 448 MB/s | 163 MB/s |
| nlohmann/json | C++ | N/A | 190 MB/s |

## Key Findings

### 1. Go can match or exceed C++ performance

The **go-json** library achieves:
- **10% faster** marshaling than RapidJSON (601 vs 548 MB/s)
- **17% slower** parsing than RapidJSON (773 vs 931 MB/s)

This makes Go a viable choice for JSON-heavy applications without significant performance penalty.

### 2. Library choice matters more than language

Within each language, library choice has a 3-4x impact:
- Go: go-json is 3.5x faster than encoding/json for parsing
- C++: RapidJSON is 4x faster than nlohmann/json for parsing

### 3. encoding/json is the bottleneck in gotio

The standard library's `encoding/json` achieves only:
- 125-163 MB/s for parsing (vs 773 MB/s with go-json)
- 447 MB/s for marshaling (vs 601 MB/s with go-json)

**Recommendation:** Switch gotio to use go-json for 2-5x performance improvement.

## Libraries Tested

### Go
- **encoding/json** (stdlib) - Reflection-based, widely compatible
- **jsoniter** - Drop-in compatible, good performance
- **go-json** (goccy) - Drop-in compatible, best performance

### C++
- **RapidJSON** - SAX/DOM parser, used by OTIO reference
- **nlohmann/json** - Header-only, modern API, slower

## Running the Benchmarks

### Quick Start
```bash
cd benchmarks
./run_comparison.sh
```

### Custom Configuration
```bash
ITERATIONS=100 VIDEO_TRACKS=5 CLIPS=200 ./run_comparison.sh
```

### Individual Benchmarks

**Go:**
```bash
cd go-json-bench
go build -o json-benchmark .
./json-benchmark --iterations 100 --video-tracks 5 --clips 200
```

**C++:**
```bash
cd cpp-json-bench
mkdir -p build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j8
./json-benchmark --iterations 100 --video-tracks 5 --clips 200
```

### Generating Test Data
```bash
./go-json-bench/json-benchmark --generate testdata --generate-count 20
```

## Test Data Structure

The benchmarks use OTIO-like timeline structures:
- Timeline with configurable video/audio tracks
- Clips with source ranges, media references, metadata
- Nested JSON matching real OTIO files

Sample sizes generated:
- Small: 36 KB (1 video + 1 audio, 10 clips)
- Medium: 356 KB (2 video + 2 audio, 50 clips)
- Standard: 887 KB (3 video + 2 audio, 100 clips)
- Large: 3.2 MB (5 video + 4 audio, 200 clips)
- XLarge: 16 MB (10 video + 8 audio, 500 clips)

## Integration Findings for gotio

### Attempted Integration

We attempted to integrate both go-json and jsoniter into gotio as drop-in replacements for encoding/json.

**Result: Both were SLOWER, not faster.**

| Library | RationalTime Marshal | Timeline Marshal |
|---------|---------------------|------------------|
| encoding/json | 190 ns | 363 us |
| jsoniter | 1,947 ns (10x slower) | ~5 ms (14x slower) |
| go-json | 1,530 ns (8x slower) | ~1.5 ms (4x slower) |

### Why the Discrepancy?

The isolated benchmarks (in this directory) show go-json/jsoniter performing well because they test **direct struct marshaling**. However, gotio uses **custom MarshalJSON/UnmarshalJSON methods** that internally call `json.Marshal` on nested structs.

This recursive pattern creates overhead in the alternative libraries:
1. Extra reflection/type resolution on each nested call
2. Loss of optimization opportunities that encoding/json has for simple structs
3. Additional allocations for interface boxing

### Recommendation

**Keep encoding/json** for gotio's current architecture. The custom MarshalJSON pattern is incompatible with drop-in JSON library replacements.

For significant JSON performance gains, consider:
1. **Code generation with easyjson** - generates static marshaling code
2. **Streaming JSON** - use jsoniter's Iterator API for large files
3. **Architectural change** - marshal entire object graphs without custom methods

## Raw Benchmark Output

### Go (Apple M3 Max)
```
Library              Operation              Throughput  Avg Latency
go-json              Marshal                 601.05 MB/s    726.94 us
encoding/json        Marshal                 447.35 MB/s    976.72 us
jsoniter             Marshal                 395.07 MB/s   1105.94 us

go-json              Unmarshal               445.68 MB/s    980.36 us
jsoniter             Unmarshal               296.35 MB/s   1474.36 us
encoding/json        Unmarshal               124.95 MB/s   3496.98 us

go-json              Unmarshal (files)       773.32 MB/s   5041.50 us
jsoniter             Unmarshal (files)       504.63 MB/s   7725.80 us
encoding/json        Unmarshal (files)       163.06 MB/s  23909.03 us
```

### C++ (Apple M3 Max)
```
Library             Operation             Throughput Avg Latency
RapidJSON           Stringify              548.18 MB/s   810.82 us
nlohmann/json       Stringify              328.61 MB/s  1352.60 us

RapidJSON           Parse                  539.10 MB/s   824.48 us
nlohmann/json       Parse                  135.74 MB/s  3274.58 us

RapidJSON           Parse (files)          930.66 MB/s  4189.14 us
nlohmann/json       Parse (files)          189.92 MB/s 20527.76 us
```
