// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

import (
	"bytes"
	"encoding/json"
)

// RawMessage is a raw encoded JSON value used for polymorphic fields.
type RawMessage = json.RawMessage

// jsonIndent formats JSON with indentation.
func jsonIndent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return json.Indent(dst, src, prefix, indent)
}
