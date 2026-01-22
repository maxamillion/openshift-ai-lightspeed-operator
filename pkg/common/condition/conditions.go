/*
Copyright 2022 Red Hat
Copyright 2025.

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

// +kubebuilder:object:generate:=true

package condition

// Common Condition Types used by API objects.
const (
	// ReadyCondition defines the Ready condition type that summarizes the operational state of the object.
	ReadyCondition Type = "Ready"
)

// Common Reasons used by API objects.
const (
	// RequestedReason (Alarm) documents a condition not in Status=True because creation has been requested
	RequestedReason Reason = "Requested"

	// InitReason documents a condition not in Status=True because the underlying object is in init state
	InitReason Reason = "Init"

	// ReadyReason documents a condition is in Status=True because the underlying object is ready
	ReadyReason Reason = "Ready"

	// ErrorReason (Alarm) documents a condition not in Status=True because the underlying object
	// encountered an error
	ErrorReason Reason = "Error"
)

// Common messages used by API objects
const (
	// ReadyMessage defines the message shown when the object is ready
	ReadyMessage = "Setup complete"

	// ReadyInitMessage defines the initial message shown when the object is not yet ready
	ReadyInitMessage = "Setup not started"

	// DeploymentReadyErrorMessage defines the message shown when deployment fails
	DeploymentReadyErrorMessage = "Deployment not ready: %s"
)
