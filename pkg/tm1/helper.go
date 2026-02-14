package tm1

import (
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// CompareTM1Versions compares two dotted TM1 version strings.
// Returns 1 if v1 > v2, -1 if v1 < v2, and 0 if equal.
func CompareTM1Versions(v1, v2 string) int {
	segments1 := strings.Split(strings.TrimSpace(v1), ".")
	segments2 := strings.Split(strings.TrimSpace(v2), ".")

	for len(segments1) < len(segments2) {
		segments1 = append(segments1, "0")
	}
	for len(segments2) < len(segments1) {
		segments2 = append(segments2, "0")
	}

	for i := 0; i < len(segments1); i++ {
		seg1, _ := strconv.Atoi(segments1[i])
		seg2, _ := strconv.Atoi(segments2[i])
		if seg1 > seg2 {
			return 1
		}
		if seg1 < seg2 {
			return -1
		}
	}

	return 0
}

// IsV1GreaterOrEqualToV2 returns true when v1 is greater than or equal to v2.
func IsV1GreaterOrEqualToV2(v1, v2 string) bool {
	return CompareTM1Versions(v1, v2) >= 0
}

// TM1MajorVersion extracts the major version number from a TM1 version string.
func TM1MajorVersion(version string) (int, bool) {
	version = strings.TrimSpace(version)
	if version == "" {
		return 0, false
	}

	segments := strings.Split(version, ".")
	if len(segments) == 0 {
		return 0, false
	}

	major, err := strconv.Atoi(segments[0])
	if err != nil {
		return 0, false
	}

	return major, true
}

type CellUpdateableProperty int

const (
	SECURITY_RESTRICTED CellUpdateableProperty = iota + 1
	UPDATE_CUBE_APPLICABLE
	RULE_IS_APPLIED
	PICKLIST_EXISTS
	SANDBOX_VALUE_IS_DIFFERENT_TO_BASE
	_
	_
	_
	NO_SPREADING_HOLD
	LEAF_HOLD
	CONSOLIDATION_SPREADING_HOLD
	TEMPORARY_SPREADING_HOLD
	_
	_
	CELL_IS_NOT_UPDATEABLE = 29
)

// ExtractCellUpdateableProperty extracts a specific property bit from a decimal value.
func ExtractCellUpdateableProperty(decimalValue int, cellProperty CellUpdateableProperty) bool {
	bit := (decimalValue & (1 << (cellProperty - 1))) != 0
	return bit
}

// CellIsUpdateable checks if a cell is updateable.
func CellIsUpdateable(cell Cell) bool {
	bit := ExtractCellUpdateableProperty(cell.Updateable, CELL_IS_NOT_UPDATEABLE)
	updateable := !bit
	return updateable
}

// RandomString generates a random string of length n using a specified character set.
func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().UnixNano())
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// AddURLParameters appends non-empty query params to a base URL.
func AddURLParameters(baseURL string, params map[string]string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	queryValues := u.Query()
	for key, value := range params {
		if value != "" {
			queryValues.Set(key, value)
		}
	}
	if u.RawQuery != "" {
		u.RawQuery += "&"
	}

	u.RawQuery += queryValues.Encode()
	return u.String(), nil
}

// EncodeODataQuery encodes OData query parameters, preserving spaces as %20.
func EncodeODataQuery(values url.Values) string {
	if values == nil {
		return ""
	}
	encoded := values.Encode()
	return strings.ReplaceAll(encoded, "+", "%20")
}

// SliceContains checks if a slice contains a value.
func SliceContains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// UniqueStrings returns a slice of unique strings from the input slice.
func UniqueStrings(input []string) []string {
	unique := make(map[string]bool)
	var result []string
	for _, item := range input {
		if _, value := unique[item]; !value {
			unique[item] = true
			result = append(result, item)
		}
	}
	return result
}

// ExtractDimensionHierarchyFromString extracts the dimension and hierarchy from a string in the format
// [dimension].[hierarchy] or dimension:hierarchy.
func ExtractDimensionHierarchyFromString(input string) (string, string) {
	var sections []string
	start := -1

	for i, char := range input {
		switch char {
		case '[':
			if start == -1 {
				start = i + 1
			}
		case ']':
			if start != -1 && i > start {
				sections = append(sections, input[start:i])
				start = -1
			}
		case '.':
			if start != -1 {
				return input, input
			}
		}
	}

	if len(sections) == 2 {
		return sections[0], sections[1]
	} else if len(sections) == 0 && strings.Contains(input, ":") {
		parts := strings.SplitN(input, ":", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
		return parts[0], ""
	}

	return input, input
}
