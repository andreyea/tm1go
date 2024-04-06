package tm1go

import (
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func isV1GreaterOrEqualToV2(v1, v2 string) bool {
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

func SliceContains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
