cmake_minimum_required(VERSION 3.14)
project(json-benchmark CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_BUILD_TYPE Release)

# Optimization flags
set(CMAKE_CXX_FLAGS_RELEASE "-O3 -DNDEBUG")

# Set minimum cmake policy for FetchContent deps
cmake_policy(SET CMP0077 NEW)

include(FetchContent)

# Fetch RapidJSON (header-only, skip building examples/tests)
FetchContent_Declare(
    rapidjson
    GIT_REPOSITORY https://github.com/Tencent/rapidjson.git
    GIT_TAG master
    GIT_SHALLOW TRUE
)
set(RAPIDJSON_BUILD_DOC OFF CACHE BOOL "" FORCE)
set(RAPIDJSON_BUILD_EXAMPLES OFF CACHE BOOL "" FORCE)
set(RAPIDJSON_BUILD_TESTS OFF CACHE BOOL "" FORCE)
FetchContent_Populate(rapidjson)

# Fetch nlohmann/json
FetchContent_Declare(
    nlohmann_json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.3
    GIT_SHALLOW TRUE
)
FetchContent_MakeAvailable(nlohmann_json)

# Optional: Fetch simdjson
option(USE_SIMDJSON "Use simdjson for parsing benchmarks" OFF)
if(USE_SIMDJSON)
    FetchContent_Declare(
        simdjson
        GIT_REPOSITORY https://github.com/simdjson/simdjson.git
        GIT_TAG v3.6.3
        GIT_SHALLOW TRUE
    )
    FetchContent_MakeAvailable(simdjson)
endif()

# Main benchmark executable
add_executable(json-benchmark main.cpp)

target_include_directories(json-benchmark PRIVATE
    ${rapidjson_SOURCE_DIR}/include
)

target_link_libraries(json-benchmark PRIVATE
    nlohmann_json::nlohmann_json
)

if(USE_SIMDJSON)
    target_compile_definitions(json-benchmark PRIVATE USE_SIMDJSON)
    target_link_libraries(json-benchmark PRIVATE simdjson)
endif()

# Print configuration
message(STATUS "Building JSON benchmark with:")
message(STATUS "  RapidJSON: ${rapidjson_SOURCE_DIR}")
message(STATUS "  nlohmann/json: enabled")
message(STATUS "  simdjson: ${USE_SIMDJSON}")
