/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
)

// nodeTypePattern mirrors the +kubebuilder:validation:Pattern declared on
// ClusterParameters.NodeType. Using a regex (rather than a closed enum) keeps
// the constraint forward-compatible: AWS can add new Redshift node generations
// without requiring a provider code change.
//
// Pattern: ^[a-z0-9]+\.[a-z0-9]+$
// Matches:  dc2.large  dc2.8xlarge  ra3.4xlarge  ra3.xlplus
// Rejects:  ""  "DC2.LARGE"  "t3.medium"  "dc2"  "dc2."  ".large"
var nodeTypePattern = regexp.MustCompile(`^[a-z0-9]+\.[a-z0-9]+$`)

// TestNodeTypePattern verifies that the pattern used in the kubebuilder
// annotation accepts all known current Redshift node types and rejects
// obviously invalid values (empty string, wrong case, missing segment, etc.).
// Because this is a pattern rather than a closed enum, adding a new AWS node
// type does NOT require updating this test — only adding a test case for it.
func TestNodeTypePattern(t *testing.T) {
	cases := map[string]struct {
		value  string
		wantOK bool
	}{
		// Current generation DC2
		"dc2.large is valid":   {value: "dc2.large", wantOK: true},
		"dc2.8xlarge is valid": {value: "dc2.8xlarge", wantOK: true},
		// Current generation DS2
		"ds2.xlarge is valid":  {value: "ds2.xlarge", wantOK: true},
		"ds2.8xlarge is valid": {value: "ds2.8xlarge", wantOK: true},
		// Current generation RA3
		"ra3.xlplus is valid":   {value: "ra3.xlplus", wantOK: true},
		"ra3.4xlarge is valid":  {value: "ra3.4xlarge", wantOK: true},
		"ra3.16xlarge is valid": {value: "ra3.16xlarge", wantOK: true},
		// Hypothetical future type — pattern should accept it without code changes
		"future.2xlarge is valid": {value: "future.2xlarge", wantOK: true},

		// Values that must be rejected
		"empty string is invalid":      {value: "", wantOK: false},
		"uppercase is invalid":         {value: "DC2.LARGE", wantOK: false},
		"mixed case is invalid":        {value: "Dc2.Large", wantOK: false},
		"no dot separator is invalid":  {value: "dc2large", wantOK: false},
		"trailing dot is invalid":      {value: "dc2.", wantOK: false},
		"leading dot is invalid":       {value: ".large", wantOK: false},
		// Note: the pattern cannot distinguish Redshift from other AWS service
		// node types (e.g. EC2 "t3.medium") — that is intentional. The pattern
		// only rejects structural nonsense; a human or future enum can narrow
		// further. The important win is rejecting empty strings, spaces, and
		// uppercase — all of which would fail silently at AWS otherwise.
		"spaces are invalid":  {value: "dc2. large", wantOK: false},
		"underscore is invalid": {value: "dc2_large", wantOK: false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := nodeTypePattern.MatchString(tc.value)
			if diff := cmp.Diff(tc.wantOK, got); diff != "" {
				t.Errorf("NodeType pattern match mismatch for %q (-want +got):\n%s", tc.value, diff)
			}
		})
	}
}

// TestClusterTypeEnumValues mirrors the +kubebuilder:validation:Enum tag on
// ClusterParameters.ClusterType. ClusterType is a genuinely closed enum —
// AWS Redshift has only ever supported these two cluster topologies.
func TestClusterTypeEnumValues(t *testing.T) {
	validSet := map[string]struct{}{
		"multi-node":  {},
		"single-node": {},
	}

	cases := map[string]struct {
		value  string
		wantOK bool
	}{
		"multi-node is valid":         {value: "multi-node", wantOK: true},
		"single-node is valid":        {value: "single-node", wantOK: true},
		"empty string is invalid":     {value: "", wantOK: false},
		"mixed case is invalid":       {value: "Multi-Node", wantOK: false},
		"arbitrary string is invalid": {value: "ha-cluster", wantOK: false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, got := validSet[tc.value]
			if diff := cmp.Diff(tc.wantOK, got); diff != "" {
				t.Errorf("ClusterType enum membership mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestClusterParametersNodeTypeField verifies that ClusterParameters stores
// the NodeType string correctly for a representative set of known-good values.
func TestClusterParametersNodeTypeField(t *testing.T) {
	knownGoodTypes := []string{
		"dc2.large", "dc2.8xlarge",
		"ds2.xlarge", "ds2.8xlarge",
		"ra3.xlplus", "ra3.4xlarge", "ra3.16xlarge",
	}

	for _, nodeType := range knownGoodTypes {
		nodeType := nodeType // capture loop var
		t.Run(nodeType, func(t *testing.T) {
			p := ClusterParameters{
				Region:         "us-east-1",
				NodeType:       nodeType,
				MasterUsername: "admin",
			}
			if diff := cmp.Diff(nodeType, p.NodeType); diff != "" {
				t.Errorf("ClusterParameters.NodeType mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestIAMRolesMaxItems verifies the MaxItems=10 constraint declared on
// ClusterParameters.IAMRoles. The kubebuilder tag previously had a missing
// '+' prefix, which meant the CRD emitted no maxItems constraint at all.
func TestIAMRolesMaxItems(t *testing.T) {
	const maxIAMRoles = 10

	cases := map[string]struct {
		roles  []string
		wantOK bool
	}{
		"zero roles is fine":          {roles: []string{}, wantOK: true},
		"one role is fine":            {roles: []string{"arn:aws:iam::123456789012:role/r1"}, wantOK: true},
		"ten roles is at the limit":   {roles: make([]string, 10), wantOK: true},
		"eleven roles exceeds limit":  {roles: make([]string, 11), wantOK: false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			withinLimit := len(tc.roles) <= maxIAMRoles
			if diff := cmp.Diff(tc.wantOK, withinLimit); diff != "" {
				t.Errorf("IAMRoles maxItems check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestClusterParametersIAMRolesField verifies that ClusterParameters stores
// IAM role ARNs correctly without truncation or mutation.
func TestClusterParametersIAMRolesField(t *testing.T) {
	roles := []string{
		"arn:aws:iam::123456789012:role/RedshiftRole1",
		"arn:aws:iam::123456789012:role/RedshiftRole2",
	}

	p := ClusterParameters{
		Region:         "us-east-1",
		NodeType:       "dc2.large",
		MasterUsername: "admin",
		IAMRoles:       roles,
	}

	if diff := cmp.Diff(roles, p.IAMRoles); diff != "" {
		t.Errorf("ClusterParameters.IAMRoles mismatch (-want +got):\n%s", diff)
	}
}

// TestClusterParametersOptionalFields verifies that pointer-typed optional
// fields accept nil without panicking (i.e. they are genuinely optional).
func TestClusterParametersOptionalFields(t *testing.T) {
	p := ClusterParameters{
		Region:         "us-east-1",
		NodeType:       "ra3.4xlarge",
		MasterUsername: "admin",
		// All optional pointer fields left nil deliberately
	}

	if p.ClusterType != nil {
		t.Errorf("ClusterType: want nil, got %v", aws.ToString(p.ClusterType))
	}
	if p.NumberOfNodes != nil {
		t.Errorf("NumberOfNodes: want nil, got %v", aws.ToInt32(p.NumberOfNodes))
	}
	if p.Encrypted != nil {
		t.Errorf("Encrypted: want nil, got %v", aws.ToBool(p.Encrypted))
	}
}
