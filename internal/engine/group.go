package engine

import (
	"path"
	"strings"
)

// UnitFor maps a file path to its indivisible directory unit. Files at the
// unit boundary are rejected so rclone copy can never reinterpret a file as a
// destination directory.
func UnitFor(relative, strategy string, depth int) string {
	relative = strings.ReplaceAll(relative, `\\`, "/")
	if strings.HasPrefix(relative, "/") {
		return ""
	}
	clean := path.Clean(relative)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return ""
	}
	parts := strings.Split(strings.Trim(clean, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	n := 1
	switch strategy {
	case "season":
		n = 2
	case "depth":
		if depth < 1 {
			return ""
		}
		n = depth
	case "show", "folder":
		n = 1
	default:
		return ""
	}
	if len(parts) <= n {
		return ""
	}
	return strings.Join(parts[:n], "/")
}

// PickDestination is deterministic, so a unit stays pinned even before persistence.
func PickDestination(unit string, weights []int) int {
	total := 0
	for _, w := range weights {
		total += max(1, w)
	}
	if total == 0 {
		return 0
	}
	var h uint32 = 2166136261
	for i := 0; i < len(unit); i++ {
		h ^= uint32(unit[i])
		h *= 16777619
	}
	n := int(h % uint32(total))
	for i, w := range weights {
		n -= max(1, w)
		if n < 0 {
			return i
		}
	}
	return 0
}
