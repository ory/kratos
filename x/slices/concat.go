// copied from https://cs.opensource.google/go/go/+/refs/tags/go1.22.0:src/slices/slices.go;l=501-515;drc=2551fffd2c06cf0655ebbbd11d9b1e70a5b2e9cb
// Delete when Kratos moves to Go 1.22

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package slices defines various functions useful with slices of any type.
package slices

import slices_ "slices"

// Concat returns a new slice concatenating the passed in slices.
func Concat[S ~[]E, E any](slices ...S) S {
	size := 0
	for _, s := range slices {
		size += len(s)
		if size < 0 {
			panic("len out of range")
		}
	}
	newslice := slices_.Grow[S](nil, size)
	for _, s := range slices {
		newslice = append(newslice, s...)
	}
	return newslice
}
