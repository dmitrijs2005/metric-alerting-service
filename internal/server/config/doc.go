// Package config provides functionality for loading application configuration
// from multiple sources with a defined priority.
//
// The configuration resolution order is:
//  1. Built-in defaults (lowest priority).
//  2. JSON configuration file, specified either by the CONFIG environment
//     variable or the -c / -config command-line flags.
//  3. Command-line flags
//  4. Environment variables (highest priority).
//
// This layered approach allows stable defaults and repeatable deployments
// (via JSON), while still supporting overrides for container environments
// (via environment variables) and quick ad-hoc testing (via CLI flags).
//
// Internally, the package uses an intermediate JsonConfig structure to parse
// JSON files. This struct uses common.Duration wrappers for duration fields
// so that values can be specified either as human-readable strings like "5s"
// or as raw integers representing nanoseconds. After parsing, the values are
// copied into the runtime Config struct, which uses time.Duration.
package config
