package provider

import (
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NameResourceModel Name module for names
type NameResourceModel struct {
	Name types.String `tfsdk:"name"`
}

// ExpontentialBackoff is a function that takes in a sleepTime and maxSleepTime and returns a new sleepTime
func ExpontentialBackoff(sleepTime int, maxSleepTime int) int {
	if sleepTime < maxSleepTime {
		sleepTime = sleepTime * 2
	}
	time.Sleep(time.Duration(sleepTime) * time.Second)
	return sleepTime
}

// StringInSlice checks if a string is in a slice of strings
func StringInSlice(str string, list []types.String) bool {
	for _, v := range list {
		if v.ValueString() == str {
			return true
		}
	}
	return false
}

// CompareVersions compares two versions and returns 1 if the first version is greater, -1 if the second version is greater, and 0 if they are equal
func CompareVersions(v1, v2 string) int {
	s1 := strings.Split(v1, ".")
	s2 := strings.Split(v2, ".")

	len1 := len(s1)
	len2 := len(s2)

	for i := 0; i < len1 || i < len2; i++ {
		part1 := 0
		if i < len1 {
			part1, _ = strconv.Atoi(s1[i])
		}

		part2 := 0
		if i < len2 {
			part2, _ = strconv.Atoi(s2[i])
		}

		if part1 > part2 {
			return 1
		} else if part1 < part2 {
			return -1
		}
	}

	return 0
}
