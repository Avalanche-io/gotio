// JSON Throughput Benchmark - C++ Implementation
// Compares RapidJSON (used by OTIO reference), nlohmann/json, and simdjson

#include <iostream>
#include <fstream>
#include <sstream>
#include <string>
#include <vector>
#include <chrono>
#include <filesystem>
#include <iomanip>
#include <algorithm>
#include <cstring>

// RapidJSON
#include "rapidjson/document.h"
#include "rapidjson/writer.h"
#include "rapidjson/stringbuffer.h"
#include "rapidjson/prettywriter.h"

// nlohmann/json
#include "nlohmann/json.hpp"

// simdjson (if available)
#ifdef USE_SIMDJSON
#include "simdjson.h"
#endif

namespace fs = std::filesystem;
using json = nlohmann::json;
using namespace std::chrono;

struct BenchmarkResult {
    std::string library;
    std::string operation;
    int iterations;
    int64_t total_bytes;
    double duration_ms;
    double throughput_mbs;
    double avg_latency_us;
};

// Generate OTIO-like timeline JSON using RapidJSON
std::string generate_timeline_json(int video_tracks, int audio_tracks, int clips_per_track) {
    rapidjson::Document doc;
    doc.SetObject();
    auto& alloc = doc.GetAllocator();

    doc.AddMember("OTIO_SCHEMA", "Timeline.1", alloc);
    doc.AddMember("name", "Benchmark Timeline", alloc);

    // Global start time
    rapidjson::Value global_start(rapidjson::kObjectType);
    global_start.AddMember("OTIO_SCHEMA", "RationalTime.1", alloc);
    global_start.AddMember("value", 86400.0, alloc);
    global_start.AddMember("rate", 24.0, alloc);
    doc.AddMember("global_start_time", global_start, alloc);

    // Tracks stack
    rapidjson::Value tracks_stack(rapidjson::kObjectType);
    tracks_stack.AddMember("OTIO_SCHEMA", "Stack.1", alloc);
    tracks_stack.AddMember("name", "tracks", alloc);

    rapidjson::Value children(rapidjson::kArrayType);

    auto create_clip = [&](int index) -> rapidjson::Value {
        rapidjson::Value clip(rapidjson::kObjectType);
        clip.AddMember("OTIO_SCHEMA", "Clip.2", alloc);

        std::string name = "Shot_" + std::to_string(index);
        rapidjson::Value name_val;
        name_val.SetString(name.c_str(), alloc);
        clip.AddMember("name", name_val, alloc);
        clip.AddMember("enabled", true, alloc);

        // Source range
        rapidjson::Value source_range(rapidjson::kObjectType);
        source_range.AddMember("OTIO_SCHEMA", "TimeRange.1", alloc);

        rapidjson::Value start_time(rapidjson::kObjectType);
        start_time.AddMember("OTIO_SCHEMA", "RationalTime.1", alloc);
        start_time.AddMember("value", static_cast<double>(index * 24), alloc);
        start_time.AddMember("rate", 24.0, alloc);
        source_range.AddMember("start_time", start_time, alloc);

        rapidjson::Value duration(rapidjson::kObjectType);
        duration.AddMember("OTIO_SCHEMA", "RationalTime.1", alloc);
        duration.AddMember("value", 48.0, alloc);
        duration.AddMember("rate", 24.0, alloc);
        source_range.AddMember("duration", duration, alloc);

        clip.AddMember("source_range", source_range, alloc);

        // Media reference
        rapidjson::Value media_ref(rapidjson::kObjectType);
        media_ref.AddMember("OTIO_SCHEMA", "ExternalReference.1", alloc);

        std::string media_name = "media_" + std::to_string(index);
        rapidjson::Value media_name_val;
        media_name_val.SetString(media_name.c_str(), alloc);
        media_ref.AddMember("name", media_name_val, alloc);

        std::string url = "file:///media/project/footage/clip_" +
                          std::string(5 - std::to_string(index).length(), '0') +
                          std::to_string(index) + ".mov";
        rapidjson::Value url_val;
        url_val.SetString(url.c_str(), alloc);
        media_ref.AddMember("target_url", url_val, alloc);

        // Available range
        rapidjson::Value avail_range(rapidjson::kObjectType);
        avail_range.AddMember("OTIO_SCHEMA", "TimeRange.1", alloc);

        rapidjson::Value avail_start(rapidjson::kObjectType);
        avail_start.AddMember("OTIO_SCHEMA", "RationalTime.1", alloc);
        avail_start.AddMember("value", 0.0, alloc);
        avail_start.AddMember("rate", 24.0, alloc);
        avail_range.AddMember("start_time", avail_start, alloc);

        rapidjson::Value avail_dur(rapidjson::kObjectType);
        avail_dur.AddMember("OTIO_SCHEMA", "RationalTime.1", alloc);
        avail_dur.AddMember("value", 1000.0, alloc);
        avail_dur.AddMember("rate", 24.0, alloc);
        avail_range.AddMember("duration", avail_dur, alloc);

        media_ref.AddMember("available_range", avail_range, alloc);

        // Metadata
        rapidjson::Value meta(rapidjson::kObjectType);
        meta.AddMember("codec", "ProRes422HQ", alloc);
        meta.AddMember("resolution", "1920x1080", alloc);
        meta.AddMember("colorspace", "Rec709", alloc);
        media_ref.AddMember("metadata", meta, alloc);

        clip.AddMember("media_reference", media_ref, alloc);

        // Clip metadata
        rapidjson::Value clip_meta(rapidjson::kObjectType);
        clip_meta.AddMember("shot_type", "wide", alloc);

        std::string scene = "Scene_" + std::to_string(index / 10);
        rapidjson::Value scene_val;
        scene_val.SetString(scene.c_str(), alloc);
        clip_meta.AddMember("scene", scene_val, alloc);

        clip_meta.AddMember("take", index % 5, alloc);
        clip_meta.AddMember("notes", "This is a sample note for the clip with some additional text to make it more realistic.", alloc);
        clip_meta.AddMember("color_tag", "green", alloc);
        clip_meta.AddMember("approved", true, alloc);
        clip_meta.AddMember("frame_rate", 24.0, alloc);
        clip.AddMember("metadata", clip_meta, alloc);

        clip.AddMember("active_media_reference_key", "DEFAULT_MEDIA", alloc);

        rapidjson::Value markers(rapidjson::kArrayType);
        clip.AddMember("markers", markers, alloc);

        rapidjson::Value effects(rapidjson::kArrayType);
        clip.AddMember("effects", effects, alloc);

        return clip;
    };

    auto create_track = [&](const std::string& name, const std::string& kind, int clip_count) -> rapidjson::Value {
        rapidjson::Value track(rapidjson::kObjectType);
        track.AddMember("OTIO_SCHEMA", "Track.1", alloc);

        rapidjson::Value name_val;
        name_val.SetString(name.c_str(), alloc);
        track.AddMember("name", name_val, alloc);

        rapidjson::Value kind_val;
        kind_val.SetString(kind.c_str(), alloc);
        track.AddMember("kind", kind_val, alloc);

        rapidjson::Value clips(rapidjson::kArrayType);
        for (int i = 0; i < clip_count; i++) {
            clips.PushBack(create_clip(i), alloc);
        }
        track.AddMember("children", clips, alloc);

        rapidjson::Value meta(rapidjson::kObjectType);
        meta.AddMember("track_index", 0, alloc);
        meta.AddMember("locked", false, alloc);
        meta.AddMember("muted", false, alloc);
        track.AddMember("metadata", meta, alloc);

        return track;
    };

    for (int i = 0; i < video_tracks; i++) {
        std::string name = "V" + std::to_string(i + 1);
        children.PushBack(create_track(name, "Video", clips_per_track), alloc);
    }

    for (int i = 0; i < audio_tracks; i++) {
        std::string name = "A" + std::to_string(i + 1);
        children.PushBack(create_track(name, "Audio", clips_per_track), alloc);
    }

    tracks_stack.AddMember("children", children, alloc);

    rapidjson::Value stack_meta(rapidjson::kObjectType);
    tracks_stack.AddMember("metadata", stack_meta, alloc);

    doc.AddMember("tracks", tracks_stack, alloc);

    rapidjson::Value timeline_meta(rapidjson::kObjectType);
    timeline_meta.AddMember("project", "Benchmark Project", alloc);
    timeline_meta.AddMember("created_by", "json-benchmark", alloc);
    doc.AddMember("metadata", timeline_meta, alloc);

    rapidjson::StringBuffer buffer;
    rapidjson::Writer<rapidjson::StringBuffer> writer(buffer);
    doc.Accept(writer);

    return buffer.GetString();
}

// RapidJSON benchmarks
BenchmarkResult benchmark_rapidjson_parse(const std::string& json_str, int iterations) {
    // Warmup
    for (int i = 0; i < 10; i++) {
        rapidjson::Document doc;
        doc.Parse(json_str.c_str());
    }

    int64_t total_bytes = static_cast<int64_t>(json_str.size()) * iterations;
    auto start = high_resolution_clock::now();

    for (int i = 0; i < iterations; i++) {
        rapidjson::Document doc;
        doc.Parse(json_str.c_str());
    }

    auto end = high_resolution_clock::now();
    double duration_ms = duration_cast<microseconds>(end - start).count() / 1000.0;
    double throughput = static_cast<double>(total_bytes) / (duration_ms / 1000.0) / (1024 * 1024);
    double avg_latency = (duration_ms * 1000.0) / iterations;

    return {"RapidJSON", "Parse", iterations, total_bytes, duration_ms, throughput, avg_latency};
}

BenchmarkResult benchmark_rapidjson_stringify(const std::string& json_str, int iterations) {
    // Parse once
    rapidjson::Document doc;
    doc.Parse(json_str.c_str());

    // Warmup
    for (int i = 0; i < 10; i++) {
        rapidjson::StringBuffer buffer;
        rapidjson::Writer<rapidjson::StringBuffer> writer(buffer);
        doc.Accept(writer);
    }

    int64_t total_bytes = 0;
    auto start = high_resolution_clock::now();

    for (int i = 0; i < iterations; i++) {
        rapidjson::StringBuffer buffer;
        rapidjson::Writer<rapidjson::StringBuffer> writer(buffer);
        doc.Accept(writer);
        total_bytes += buffer.GetSize();
    }

    auto end = high_resolution_clock::now();
    double duration_ms = duration_cast<microseconds>(end - start).count() / 1000.0;
    double throughput = static_cast<double>(total_bytes) / (duration_ms / 1000.0) / (1024 * 1024);
    double avg_latency = (duration_ms * 1000.0) / iterations;

    return {"RapidJSON", "Stringify", iterations, total_bytes, duration_ms, throughput, avg_latency};
}

// nlohmann/json benchmarks
BenchmarkResult benchmark_nlohmann_parse(const std::string& json_str, int iterations) {
    // Warmup
    for (int i = 0; i < 10; i++) {
        auto doc = json::parse(json_str);
    }

    int64_t total_bytes = static_cast<int64_t>(json_str.size()) * iterations;
    auto start = high_resolution_clock::now();

    for (int i = 0; i < iterations; i++) {
        auto doc = json::parse(json_str);
    }

    auto end = high_resolution_clock::now();
    double duration_ms = duration_cast<microseconds>(end - start).count() / 1000.0;
    double throughput = static_cast<double>(total_bytes) / (duration_ms / 1000.0) / (1024 * 1024);
    double avg_latency = (duration_ms * 1000.0) / iterations;

    return {"nlohmann/json", "Parse", iterations, total_bytes, duration_ms, throughput, avg_latency};
}

BenchmarkResult benchmark_nlohmann_stringify(const std::string& json_str, int iterations) {
    // Parse once
    auto doc = json::parse(json_str);

    // Warmup
    for (int i = 0; i < 10; i++) {
        std::string output = doc.dump();
    }

    int64_t total_bytes = 0;
    auto start = high_resolution_clock::now();

    for (int i = 0; i < iterations; i++) {
        std::string output = doc.dump();
        total_bytes += output.size();
    }

    auto end = high_resolution_clock::now();
    double duration_ms = duration_cast<microseconds>(end - start).count() / 1000.0;
    double throughput = static_cast<double>(total_bytes) / (duration_ms / 1000.0) / (1024 * 1024);
    double avg_latency = (duration_ms * 1000.0) / iterations;

    return {"nlohmann/json", "Stringify", iterations, total_bytes, duration_ms, throughput, avg_latency};
}

#ifdef USE_SIMDJSON
BenchmarkResult benchmark_simdjson_parse(const std::string& json_str, int iterations) {
    simdjson::dom::parser parser;

    // Warmup
    for (int i = 0; i < 10; i++) {
        auto doc = parser.parse(json_str);
    }

    int64_t total_bytes = static_cast<int64_t>(json_str.size()) * iterations;
    auto start = high_resolution_clock::now();

    for (int i = 0; i < iterations; i++) {
        auto doc = parser.parse(json_str);
    }

    auto end = high_resolution_clock::now();
    double duration_ms = duration_cast<microseconds>(end - start).count() / 1000.0;
    double throughput = static_cast<double>(total_bytes) / (duration_ms / 1000.0) / (1024 * 1024);
    double avg_latency = (duration_ms * 1000.0) / iterations;

    return {"simdjson", "Parse", iterations, total_bytes, duration_ms, throughput, avg_latency};
}
#endif

std::vector<std::string> load_test_files(const std::string& dir) {
    std::vector<std::string> files;
    for (const auto& entry : fs::directory_iterator(dir)) {
        if (entry.path().extension() == ".json") {
            std::ifstream file(entry.path());
            std::stringstream buffer;
            buffer << file.rdbuf();
            files.push_back(buffer.str());
        }
    }
    return files;
}

void generate_test_files(const std::string& dir, int count) {
    fs::create_directories(dir);

    struct Config {
        int video, audio, clips;
        std::string name;
    };

    std::vector<Config> configs = {
        {1, 1, 10, "small"},
        {2, 2, 50, "medium"},
        {3, 2, 100, "standard"},
        {5, 4, 200, "large"},
        {10, 8, 500, "xlarge"}
    };

    for (int i = 0; i < count; i++) {
        auto& cfg = configs[i % configs.size()];
        std::string json = generate_timeline_json(cfg.video, cfg.audio, cfg.clips);

        // Pretty print for file storage
        rapidjson::Document doc;
        doc.Parse(json.c_str());

        rapidjson::StringBuffer buffer;
        rapidjson::PrettyWriter<rapidjson::StringBuffer> writer(buffer);
        doc.Accept(writer);

        std::string filename = dir + "/timeline_" + cfg.name + "_" +
                               std::string(3 - std::to_string(i).length(), '0') +
                               std::to_string(i) + ".json";

        std::ofstream file(filename);
        file << buffer.GetString();
        file.close();

        std::cout << "  Generated " << filename << " (" << buffer.GetSize() << " bytes)" << std::endl;
    }
}

void print_results(const std::vector<BenchmarkResult>& results) {
    std::cout << "\n" << std::string(80, '=') << std::endl;
    std::cout << "BENCHMARK RESULTS" << std::endl;
    std::cout << std::string(80, '=') << std::endl;

    std::cout << std::left << std::setw(20) << "Library"
              << std::setw(20) << "Operation"
              << std::right << std::setw(12) << "Throughput"
              << std::setw(12) << "Avg Latency"
              << std::setw(12) << "Total MB" << std::endl;
    std::cout << std::string(80, '-') << std::endl;

    // Sort by operation then throughput
    auto sorted_results = results;
    std::sort(sorted_results.begin(), sorted_results.end(),
              [](const BenchmarkResult& a, const BenchmarkResult& b) {
                  if (a.operation != b.operation) return a.operation < b.operation;
                  return a.throughput_mbs > b.throughput_mbs;
              });

    std::string current_op;
    for (const auto& r : sorted_results) {
        if (r.operation != current_op) {
            if (!current_op.empty()) {
                std::cout << std::string(80, '-') << std::endl;
            }
            current_op = r.operation;
        }
        std::cout << std::left << std::setw(20) << r.library
                  << std::setw(20) << r.operation
                  << std::right << std::fixed << std::setprecision(2)
                  << std::setw(9) << r.throughput_mbs << " MB/s"
                  << std::setw(9) << r.avg_latency_us << " us"
                  << std::setw(9) << static_cast<double>(r.total_bytes) / (1024 * 1024) << " MB"
                  << std::endl;
    }
    std::cout << std::string(80, '=') << std::endl;
}

int main(int argc, char* argv[]) {
    int video_tracks = 3;
    int audio_tracks = 2;
    int clips_per_track = 100;
    int iterations = 100;
    std::string testdata_dir;
    std::string generate_dir;
    int generate_count = 10;

    // Parse arguments
    for (int i = 1; i < argc; i++) {
        std::string arg = argv[i];
        if (arg == "--video-tracks" && i + 1 < argc) {
            video_tracks = std::stoi(argv[++i]);
        } else if (arg == "--audio-tracks" && i + 1 < argc) {
            audio_tracks = std::stoi(argv[++i]);
        } else if (arg == "--clips" && i + 1 < argc) {
            clips_per_track = std::stoi(argv[++i]);
        } else if (arg == "--iterations" && i + 1 < argc) {
            iterations = std::stoi(argv[++i]);
        } else if (arg == "--testdata" && i + 1 < argc) {
            testdata_dir = argv[++i];
        } else if (arg == "--generate" && i + 1 < argc) {
            generate_dir = argv[++i];
        } else if (arg == "--generate-count" && i + 1 < argc) {
            generate_count = std::stoi(argv[++i]);
        }
    }

    // Generate mode
    if (!generate_dir.empty()) {
        std::cout << "Generating " << generate_count << " test files to " << generate_dir << std::endl;
        generate_test_files(generate_dir, generate_count);
        return 0;
    }

    std::cout << "C++ JSON Throughput Benchmark" << std::endl;
    std::cout << "=============================" << std::endl;
#if defined(__APPLE__)
    std::cout << "Platform: macOS" << std::endl;
#elif defined(__linux__)
    std::cout << "Platform: Linux" << std::endl;
#elif defined(_WIN32)
    std::cout << "Platform: Windows" << std::endl;
#endif
    std::cout << "C++ Standard: " << __cplusplus << std::endl;

    std::vector<BenchmarkResult> results;

    // Generate test data
    std::cout << "\nGenerating timeline: " << video_tracks << " video + "
              << audio_tracks << " audio tracks, " << clips_per_track << " clips each" << std::endl;

    std::string json_data = generate_timeline_json(video_tracks, audio_tracks, clips_per_track);
    std::cout << "Timeline JSON size: " << std::fixed << std::setprecision(2)
              << static_cast<double>(json_data.size()) / (1024 * 1024) << " MB" << std::endl;
    std::cout << "Running " << iterations << " iterations per library" << std::endl;

    // RapidJSON benchmarks
    std::cout << "\nBenchmarking RapidJSON..." << std::endl;
    auto result = benchmark_rapidjson_stringify(json_data, iterations);
    results.push_back(result);
    std::cout << "  Stringify: " << result.throughput_mbs << " MB/s" << std::endl;

    result = benchmark_rapidjson_parse(json_data, iterations);
    results.push_back(result);
    std::cout << "  Parse: " << result.throughput_mbs << " MB/s" << std::endl;

    // nlohmann/json benchmarks
    std::cout << "\nBenchmarking nlohmann/json..." << std::endl;
    result = benchmark_nlohmann_stringify(json_data, iterations);
    results.push_back(result);
    std::cout << "  Stringify: " << result.throughput_mbs << " MB/s" << std::endl;

    result = benchmark_nlohmann_parse(json_data, iterations);
    results.push_back(result);
    std::cout << "  Parse: " << result.throughput_mbs << " MB/s" << std::endl;

#ifdef USE_SIMDJSON
    // simdjson benchmarks (parse only - simdjson is read-only)
    std::cout << "\nBenchmarking simdjson..." << std::endl;
    result = benchmark_simdjson_parse(json_data, iterations);
    results.push_back(result);
    std::cout << "  Parse: " << result.throughput_mbs << " MB/s" << std::endl;
#endif

    // File benchmarks
    if (!testdata_dir.empty()) {
        std::cout << "\nLoading test files from " << testdata_dir << std::endl;
        auto files = load_test_files(testdata_dir);
        if (!files.empty()) {
            int64_t total_size = 0;
            for (const auto& f : files) {
                total_size += f.size();
            }
            std::cout << "Loaded " << files.size() << " files, total "
                      << static_cast<double>(total_size) / (1024 * 1024) << " MB" << std::endl;

            // Run file benchmarks for each library
            int file_iterations = iterations / 10;

            // RapidJSON file benchmarks
            {
                int64_t parse_bytes = 0;
                auto start = high_resolution_clock::now();
                for (int iter = 0; iter < file_iterations; iter++) {
                    for (const auto& f : files) {
                        rapidjson::Document doc;
                        doc.Parse(f.c_str());
                        parse_bytes += f.size();
                    }
                }
                auto end = high_resolution_clock::now();
                double duration_ms = duration_cast<microseconds>(end - start).count() / 1000.0;
                results.push_back({
                    "RapidJSON", "Parse (files)",
                    file_iterations * static_cast<int>(files.size()),
                    parse_bytes, duration_ms,
                    static_cast<double>(parse_bytes) / (duration_ms / 1000.0) / (1024 * 1024),
                    (duration_ms * 1000.0) / (file_iterations * files.size())
                });
            }

            // nlohmann file benchmarks
            {
                int64_t parse_bytes = 0;
                auto start = high_resolution_clock::now();
                for (int iter = 0; iter < file_iterations; iter++) {
                    for (const auto& f : files) {
                        auto doc = json::parse(f);
                        parse_bytes += f.size();
                    }
                }
                auto end = high_resolution_clock::now();
                double duration_ms = duration_cast<microseconds>(end - start).count() / 1000.0;
                results.push_back({
                    "nlohmann/json", "Parse (files)",
                    file_iterations * static_cast<int>(files.size()),
                    parse_bytes, duration_ms,
                    static_cast<double>(parse_bytes) / (duration_ms / 1000.0) / (1024 * 1024),
                    (duration_ms * 1000.0) / (file_iterations * files.size())
                });
            }
        }
    }

    print_results(results);

    // Summary
    std::cout << "\n" << std::string(80, '=') << std::endl;
    std::cout << "SUMMARY FOR CROSS-LANGUAGE COMPARISON" << std::endl;
    std::cout << std::string(80, '=') << std::endl;
    std::cout << "Data size: " << static_cast<double>(json_data.size()) / (1024 * 1024) << " MB" << std::endl;
    std::cout << "Iterations: " << iterations << std::endl;

    // Find best results
    BenchmarkResult best_parse{"", "Parse", 0, 0, 0, 0, 0};
    BenchmarkResult best_stringify{"", "Stringify", 0, 0, 0, 0, 0};

    for (const auto& r : results) {
        if (r.operation == "Parse" && r.throughput_mbs > best_parse.throughput_mbs) {
            best_parse = r;
        }
        if (r.operation == "Stringify" && r.throughput_mbs > best_stringify.throughput_mbs) {
            best_stringify = r;
        }
    }

    std::cout << "\nBest Parse: " << best_parse.library << " at "
              << best_parse.throughput_mbs << " MB/s (" << best_parse.avg_latency_us << " us/op)" << std::endl;
    std::cout << "Best Stringify: " << best_stringify.library << " at "
              << best_stringify.throughput_mbs << " MB/s (" << best_stringify.avg_latency_us << " us/op)" << std::endl;

    return 0;
}
