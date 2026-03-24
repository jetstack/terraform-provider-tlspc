// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strconv"
	"strings"

	"terraform-provider-tlspc/internal/tlspc"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func keyTypesFromAlgorithms(in []types.String) tlspc.KeyTypes {
	// Take in a list of allowed key algorithms and return API compatible objects.
	// Validation of input is performed at the schema by tfsdk so all inputs can be assumed to be valid.

	klength := []int32{}
	kcurves := []string{}

	for _, v := range in {
		prts := strings.Split(v.ValueString(), "_")
		// If RSA convert the length part to int32 and add to list.
		if prts[0] == "RSA" {
			// Safe to ignore error as validator ensures int.
			length, _ := strconv.Atoi(prts[1])
			klength = append(klength, int32(length))
		}
		// If EC check curve is known and add to list.
		if prts[0] == "EC" {
			kcurves = append(kcurves, prts[1])
		}
	}

	return tlspc.KeyTypes{
		RSALengths: klength,
		ECCurves:   kcurves,
	}
}

func keyAlgorithmsFromKeyTypes(in tlspc.KeyTypes) []types.String {
	// Take in a list of API key type objects and return a list of allowed key algorithms.
	out := []types.String{}

	for _, v := range in.RSALengths {
		out = append(out, types.StringValue(fmt.Sprintf("RSA_%d", v)))
	}
	for _, v := range in.ECCurves {
		out = append(out, types.StringValue(fmt.Sprintf("EC_%s", v)))
	}

	return out
}
