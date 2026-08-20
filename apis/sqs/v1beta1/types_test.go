/*
Copyright 2019 The Crossplane Authors.

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

package v1beta1

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
)

// TestFIFOQueueFieldIsBool confirms that FIFOQueue is modelled as a *bool.
// This is correct: the AWS SQS API sends "true"/"false" strings and the bool
// type prevents any other value from being submitted — there is no need for a
// string enum on this field. A nil value means the field is absent (standard
// queue default); true = FIFO; false = explicit standard.
func TestFIFOQueueFieldIsBool(t *testing.T) {
	cases := map[string]struct {
		fifoQueue *bool
		want      *bool
	}{
		"nil (standard queue — default)": {fifoQueue: nil, want: nil},
		"true (FIFO queue)":              {fifoQueue: aws.Bool(true), want: aws.Bool(true)},
		"false (explicit standard)":      {fifoQueue: aws.Bool(false), want: aws.Bool(false)},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			p := QueueParameters{
				Region:    "us-east-1",
				FIFOQueue: tc.fifoQueue,
			}
			if diff := cmp.Diff(tc.want, p.FIFOQueue); diff != "" {
				t.Errorf("QueueParameters.FIFOQueue mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestSqsManagedSseEnabledFieldIsOptional verifies that SqsManagedSseEnabled
// is a *bool (pointer, hence optional/omitempty) so it does not appear in
// serialised output when unset. The +optional marker was previously missing
// from the field; this test documents the expected nil-safe behaviour.
func TestSqsManagedSseEnabledFieldIsOptional(t *testing.T) {
	cases := map[string]struct {
		enabled *bool
		want    *bool
	}{
		"nil (field absent)":   {enabled: nil, want: nil},
		"true (SSE enabled)":   {enabled: aws.Bool(true), want: aws.Bool(true)},
		"false (SSE disabled)": {enabled: aws.Bool(false), want: aws.Bool(false)},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			p := QueueParameters{
				Region:               "us-east-1",
				SqsManagedSseEnabled: tc.enabled,
			}
			if diff := cmp.Diff(tc.want, p.SqsManagedSseEnabled); diff != "" {
				t.Errorf("QueueParameters.SqsManagedSseEnabled mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
