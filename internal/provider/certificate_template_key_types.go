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

// Keep a short list of the only 8 options supported.
var allowedAlgorithms = []string{
	"RSA_1024",
	"RSA_2048",
	"RSA_3072",
	"RSA_4096",
	"EC_P256",
	"EC_P384",
	"EC_P521",
	"EC_ED25519",
}

func keyTypesFromAlgorithms(in []types.String) []tlspc.KeyType {
	// Take in a list of allowed key algorithms and return API compatible objects.
	// Validation of input is performed at the schema by tfsdk so all inputs can be assumed to be valid.
	out := make([]tlspc.KeyType, 0, len(in))

	klength := []int32{}
	kcurves := []string{}

	for _, v := range in {
		prts := strings.Split(v.ValueString(), "_")
		// If RSA convert the length part to int32 and add to list.
		if prts[0] == "RSA" {
			length, _ := strconv.Atoi(prts[1])
			klength = append(klength, int32(length))
		}
		// If EC check curve is known and add to list.
		if prts[0] == "EC" {
			kcurves = append(kcurves, prts[1])
		}
	}

	// If we have any RSA inputs, output correct API object.
	if len(klength) > 0 {
		obj := tlspc.KeyType{
			Type:       "RSA",
			KeyLengths: klength,
		}
		out = append(out, obj)
	}

	// If we have any EC inputs, output correct API object.
	if len(kcurves) > 0 {
		obj := tlspc.KeyType{
			Type:      "EC",
			KeyCurves: kcurves,
		}
		out = append(out, obj)
	}

	return out
}

func keyAlgorithmsFromKeyTypes(in []tlspc.KeyType) []types.String {
	// Take in a list of API key type objects and return a list of allowed key algorithms.
	out := []types.String{}

	for _, v := range in {
		if v.Type == "RSA" {
			for _, l := range v.KeyLengths {
				out = append(out, types.StringValue(fmt.Sprintf("RSA_%d", l)))
			}
		}
		if v.Type == "EC" {
			for _, c := range v.KeyCurves {
				out = append(out, types.StringValue(fmt.Sprintf("EC_%s", c)))
			}
		}
	}

	return out
}
