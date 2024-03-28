package wd

import (
	"fmt"
	"sort"
	"strings"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func GoOS(id string) (string, bool) {
	if id == "darwin" || id == "windows" {
		return id, true
	}
	return "linux", true
}

func GoArch(os string, unameM string) (string, bool) {
	switch os {
	case "windows":
		v, ok := windowsArches[unameM]
		return v, ok
	case "linux":
		v, ok := linuxArches[unameM]
		return v, ok
	case "darwin":
		v, ok := darwinArches[unameM]
		return v, ok
	}
	return "", false
}

var windowsArches = map[string]string{
	"x86_64":  "amd64",
	"aarch64": "arm64",
}

var linuxArches = map[string]string{
	"x86_64":  "amd64",
	"aarch64": "arm64",
}

var darwinArches = map[string]string{
	"x86_64": "amd64",
	"arm64":  "arm64",
}

type Platform v1.Platform

func (p Platform) String() string {
	if p.OS == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString(p.OS)
	if p.Architecture != "" {
		b.WriteString("/")
		b.WriteString(p.Architecture)
	}
	if p.Variant != "" {
		b.WriteString("/")
		b.WriteString(p.Variant)
	}
	if p.OSVersion != "" {
		b.WriteString(":")
		b.WriteString(p.OSVersion)
	}
	return b.String()
}

// ParsePlatform parses a string representing a Platform, if possible.
func ParsePlatform(s string) (*Platform, error) {
	var p Platform
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) == 2 {
		p.OSVersion = parts[1]
	}
	parts = strings.Split(parts[0], "/")
	if len(parts) > 0 {
		p.OS = parts[0]
	}
	if len(parts) > 1 {
		p.Architecture = parts[1]
	}
	if len(parts) > 2 {
		p.Variant = parts[2]
	}
	if len(parts) > 3 {
		return nil, fmt.Errorf("too many slashes in platform spec: %s", s)
	}
	return &p, nil
}

// Equals returns true if the given platform is semantically equivalent to this one.
// The order of Features and OSFeatures is not important.
func (p Platform) Equals(o Platform) bool {
	return p.OS == o.OS &&
		p.Architecture == o.Architecture &&
		p.Variant == o.Variant &&
		p.OSVersion == o.OSVersion &&
		stringSliceEqualIgnoreOrder(p.OSFeatures, o.OSFeatures)
}

func stringSliceEqualIgnoreOrder(a, b []string) bool {
	if a != nil && b != nil {
		sort.Strings(a)
		sort.Strings(b)
	}
	return stringSliceEqual(a, b)
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, elm := range a {
		if elm != b[i] {
			return false
		}
	}
	return true
}
