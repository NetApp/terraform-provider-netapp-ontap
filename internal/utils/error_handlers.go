package utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ErrorHandler creates an error handler to combine logging and reporting errors
type ErrorHandler struct {
	Ctx    context.Context
	diags  *diag.Diagnostics
	name   string
	subCtx context.Context
}

// NewErrorHandler creates an error handler based on current context and TF diagnostics
func NewErrorHandler(ctx context.Context, diags *diag.Diagnostics) *ErrorHandler {
	name := "error_handler"
	return &ErrorHandler{
		Ctx:    ctx,
		diags:  diags,
		name:   name,
		subCtx: tflog.NewSubsystem(ctx, name, tflog.WithAdditionalLocationOffset(1)),
	}
}

// MakeAndLogError builds an error using message and logs the error with tflog
func (e *ErrorHandler) MakeAndLogError(msg string) error {
	e.validate()
	tflog.WithAdditionalLocationOffset(1)
	tflog.SubsystemError(e.subCtx, e.name, msg)
	return errors.New(msg)
}

// MakeAndReportError builds an error using message and logs the error with tflog
// The error is added to the diagnostic and will be reported by Terraform
func (e *ErrorHandler) MakeAndReportError(summary string, msg string) error {
	e.validate()
	fullMsg := fmt.Sprintf("HERE  %s: %s", summary, msg)
	tflog.SubsystemError(e.subCtx, e.name, msg)
	e.diags.AddError(summary, msg)
	return errors.New(fullMsg)
}

func (e *ErrorHandler) validate() {
	if e == nil {
		panic("Error handler is not set")
	}
}
