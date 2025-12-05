// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func Uuid() uuidValidator {
	return uuidValidator{}
}

type uuidValidator struct {
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v uuidValidator) Description(ctx context.Context) string {
	return "string must be a uuid"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v uuidValidator) MarkdownDescription(ctx context.Context) string {
	return "string must be a uuid"
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and updating `resp` with diagnostics.
func (v uuidValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if err := uuid.Validate(req.ConfigValue.ValueString()); err != nil {

		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid uuid",
			fmt.Sprintf("String must be a uuid: %s", err),
		)

		return
	}
}
