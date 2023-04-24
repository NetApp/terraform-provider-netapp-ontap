package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// NameResourceModel Name module for names
type NameResourceModel struct {
	Name types.String `tfsdk:"name"`
}
