package utils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringDefaultModifier is a plan modifier that sets a default value for a
// types.StringType attribute when it is not configured. The attribute must be
// marked as Optional and Computed. When setting the state during the resource
// Create, Read, or Update methods, this default value must also be included or
// the Terraform CLI will generate an error.
type StringDefaultModifier struct {
	Default string
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m StringDefaultModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %s", m.Default)
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m StringDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%s`", m.Default)
}

// Modify runs the logic of the plan modifier.
// Access to the configuration, plan, and state is available in `req`, while
// `resp` contains fields for updating the planned value, triggering resource
// replacement, and returning diagnostics.
func (m StringDefaultModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	// If the value is unknown or known, do not set default value.
	if !req.AttributePlan.IsNull() {
		return
	}

	// types.String must be the attr.Value produced by the attr.Type in the schema for this attribute
	// for generic plan modifiers, use
	// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#ConvertValue
	// to convert into a known type.
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributePlan, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	resp.AttributePlan = types.StringValue(m.Default)
}

// StringDefault is a wrapper to call the plan modifier.
func StringDefault(defaultValue string) StringDefaultModifier {
	return StringDefaultModifier{
		Default: defaultValue,
	}
}

// BoolDefaultModifier structure
type BoolDefaultModifier struct {
	Default bool
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m BoolDefaultModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %t", m.Default)
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m BoolDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%t`", m.Default)
}

// Modify runs the logic of the plan modifier.
// Access to the configuration, plan, and state is available in `req`, while
// `resp` contains fields for updating the planned value, triggering resource
// replacement, and returning diagnostics.
func (m BoolDefaultModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	// If the value is unknown or known, do not set default value.
	if !req.AttributePlan.IsNull() {
		return
	}
	var b types.Bool
	diags := tfsdk.ValueAs(ctx, req.AttributePlan, &b)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	resp.AttributePlan = types.BoolValue(m.Default)
}

// BoolDefault is a wrapper to call the plan modifier.
func BoolDefault(defaultValue bool) BoolDefaultModifier {
	return BoolDefaultModifier{
		Default: defaultValue,
	}
}

// IntDefaultModifier struct
type IntDefaultModifier struct {
	Default int64
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m IntDefaultModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %d", m.Default)
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m IntDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%d`", m.Default)
}

// Modify runs the logic of the plan modifier.
// Access to the configuration, plan, and state is available in `req`, while
// `resp` contains fields for updating the planned value, triggering resource
// replacement, and returning diagnostics.
func (m IntDefaultModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	// If the value is unknown or known, do not set default value.
	if !req.AttributePlan.IsNull() {
		return
	}
	var i types.Int64
	diags := tfsdk.ValueAs(ctx, req.AttributePlan, &i)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	resp.AttributePlan = types.Int64Value(m.Default)
}

// Int64Default is a wrapper to call the plan modifier.
func Int64Default(defaultValue int64) IntDefaultModifier {
	return IntDefaultModifier{
		Default: defaultValue,
	}
}
