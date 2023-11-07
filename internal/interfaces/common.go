package interfaces

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NameDataModel is the standard name/uuid pair that required by most resources
type NameDataModel struct {
	Name string
	UUID string
}

func StringInSlice(str string, list []types.String) bool {
	for _, v := range list {
		if v.ValueString() == str {
			return true
		}
	}
	return false
}
