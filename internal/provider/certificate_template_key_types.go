// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"terraform-provider-tlspc/internal/tlspc"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

var allowedKeyLengths = []int32{1024, 2048, 3072, 4096}
var allowedKeyCurves = []string{"P256", "P384", "P521", "ED25519"}

func keyTypesFromAlgorithms(in []types.String) []tlspc.KeyType {
	// Take in a list of allowed key algorithms and return API compatible objects.
	out := make([]tlspc.KeyType, 0, len(in))

	klength := []int32{}
	kcurves := []string{}

	for _, v := range in {
		prts := strings.Split(v.ValueString(), "_")

		// First check no unknown key types or malformed input.
		if prts[0] != "RSA" && prts[0] != "EC" || len(prts) != 2 {
			fmt.Printf("Invalid key algorithm: %s\n", v.ValueString())
			continue
		}

		// If RSA convert the length part to int32 and add to list.
		if prts[0] == "RSA" {
			length, err := strconv.Atoi(prts[1])
			if err != nil {
				fmt.Printf("Invalid key length in algorithm: %s\n", prts[1])
				continue
			}
			if !slices.Contains(allowedKeyLengths, int32(length)) {
				fmt.Printf("Unsupported key length: %d\n", length)
				continue
			}
			klength = append(klength, int32(length))
		}

		// If EC check curve is known and add to list.
		if prts[0] == "EC" {
			if !slices.Contains(allowedKeyCurves, prts[1]) {
				fmt.Printf("Unsupported key curve: %s\n", prts[1])
				continue
			}
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
