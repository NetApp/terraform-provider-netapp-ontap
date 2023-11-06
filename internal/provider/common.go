package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"time"
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
