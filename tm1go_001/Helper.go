package tm1go

import (
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func IsV1GreaterOrEqualToV2(v1, v2 string) bool {
	splitV1 := strings.Split(v1, ".")
	splitV2 := strings.Split(v2, ".")

	for len(splitV1) < len(splitV2) {
		splitV1 = append(splitV1, "0")
	}
	for len(splitV2) < len(splitV1) {
		splitV2 = append(splitV2, "0")
	}

	// Compare each segment
	for i := 0; i < len(splitV1); i++ {
		segmentV1, _ := strconv.Atoi(splitV1[i])
		segmentV2, _ := strconv.Atoi(splitV2[i])
		if segmentV1 > segmentV2 {
			return true
		} else if segmentV1 < segmentV2 {
			return false
		}
	}
	return true
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

// Function to extract a specific property bit from a decimal value
func ExtractCellUpdateableProperty(decimalValue int, cellProperty CellUpdateableProperty) bool {
	bit := (decimalValue & (1 << (cellProperty - 1))) != 0
	return bit
}

// Function to check if a cell is updateable
func CellIsUpdateable(cell Cell) bool {
	bit := ExtractCellUpdateableProperty(cell.Updateable, CELL_IS_NOT_UPDATEABLE)
	updateable := !bit
	return updateable
}

// RandomString generates a random string of length n using a specified character set
func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().UnixNano())
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

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

// SliceContains checks if a slice contains a value. It uses generics to work with any comparable type.
func SliceContains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// UniqueStrings returns a slice of unique strings from the input slice
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

// ExtractDimensionHierarchyFromString extracts the dimension and hierarchy from a string in the format [dimension].[hierarchy] or dimension:hierarchy
func ExtractDimensionHierarchyFromString(input string) (string, string) {
	var sections []string
	start := -1

	// Iterate over the string to find brackets and dot separators
	for i, char := range input {
		switch char {
		case '[':
			if start == -1 { // Ensure no nested brackets
				start = i + 1 // Mark the start after the '['
			}
		case ']':
			if start != -1 && i > start { // Ensure there's a valid '[' before ']'
				sections = append(sections, input[start:i])
				start = -1 // Reset start index after adding to sections
			}
		case '.':
			if start != -1 { // If '.' found within brackets, it's not a valid format
				return input, input
			}
		}
	}

	// Check if the format exactly matches [text].[text]
	if len(sections) == 2 {
		return sections[0], sections[1]
	} else if len(sections) == 0 && strings.Contains(input, ":") {
		// If no brackets found but input contains ":", split by ":"
		parts := strings.SplitN(input, ":", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
		return parts[0], "" // In case there is nothing after the colon
	}

	// If not matching or partially matching, return the original input twice
	return input, input
}
