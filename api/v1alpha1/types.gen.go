// Package v1alpha1 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.15.0 DO NOT EDIT.
package v1alpha1

const (
	BearerTokenScopes = "BearerToken.Scopes"
)

// Device Device represents a physical device.
type Device struct {
	// ApiVersion APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
	ApiVersion *string `json:"apiVersion,omitempty"`

	// Kind Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	Kind *string `json:"kind,omitempty"`

	// Metadata ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *ObjectMeta `json:"metadata,omitempty"`

	// Spec DeviceSpec is a description of a device's target state.
	Spec *DeviceSpec `json:"spec,omitempty"`

	// Status DeviceStatus represents information about the status of a device. Status may trail the actual state of a device, especially if the device has not contacted the management service in a while.
	Status *DeviceStatus `json:"status,omitempty"`
}

// DeviceCondition DeviceCondition contains condition information for a device.
type DeviceCondition struct {
	LastHeartbeatTime  *string `json:"lastHeartbeatTime,omitempty"`
	LastTransitionTime *string `json:"lastTransitionTime,omitempty"`

	// Message Human readable message indicating details about last transition.
	Message *string `json:"message,omitempty"`

	// Reason (brief) reason for the condition's last transition.
	Reason *string `json:"reason,omitempty"`

	// Status Status of the condition, one of True, False, Unknown.
	Status string `json:"status"`

	// Type Type of node condition.
	Type string `json:"type"`
}

// DeviceConfigSpec defines model for DeviceConfigSpec.
type DeviceConfigSpec struct {
	Inline *string `json:"inline,omitempty"`
	Name   *string `json:"name,omitempty"`
}

// DeviceList DeviceList is a list of Devices.
type DeviceList struct {
	// ApiVersion APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
	ApiVersion *string `json:"apiVersion,omitempty"`

	// Items List of pods. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	Items []Device `json:"items"`

	// Kind Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	Kind *string `json:"kind,omitempty"`

	// Metadata ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}.
	Metadata *ListMeta `json:"metadata,omitempty"`
}

// DeviceOSSpec defines model for DeviceOSSpec.
type DeviceOSSpec struct {
	// Image ostree image name or URL.
	Image string `json:"image"`
}

// DeviceSpec DeviceSpec is a description of a device's target state.
type DeviceSpec struct {
	// Config List of config resources.
	Config []DeviceConfigSpec `json:"config"`
	Os     DeviceOSSpec       `json:"os"`
}

// DeviceStatus DeviceStatus represents information about the status of a device. Status may trail the actual state of a device, especially if the device has not contacted the management service in a while.
type DeviceStatus struct {
	// Conditions Current state of the device.
	Conditions *[]DeviceCondition `json:"conditions,omitempty"`

	// SystemInfo DeviceSystemInfo is a set of ids/uuids to uniquely identify the node.
	SystemInfo *DeviceSystemInfo `json:"systemInfo,omitempty"`
}

// DeviceSystemInfo DeviceSystemInfo is a set of ids/uuids to uniquely identify the node.
type DeviceSystemInfo struct {
	// Architecture The Architecture reported by the device.
	Architecture string `json:"architecture"`

	// BootID Boot ID reported by the device.
	BootID string `json:"bootID"`

	// MachineID MachineID reported by the device.
	MachineID string `json:"machineID"`

	// OperatingSystem The Operating System reported by the device.
	OperatingSystem string `json:"operatingSystem"`
}

// ListMeta ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}.
type ListMeta struct {
	// Continue continue may be set if the user set a limit on the number of items returned, and indicates that the server has more data available. The value is opaque and may be used to issue another request to the endpoint that served this list to retrieve the next set of available objects. Continuing a consistent list may not be possible if the server configuration has changed or more than a few minutes have passed. The resourceVersion field returned when using this continue value will be identical to the value in the first response, unless you have received this token from an error message.
	Continue *string `json:"continue,omitempty"`

	// RemainingItemCount remainingItemCount is the number of subsequent items in the list which are not included in this list response. If the list request contained label or field selectors, then the number of remaining items is unknown and the field will be left unset and omitted during serialization. If the list is complete (either because it is not chunking or because this is the last chunk), then there are no more remaining items and this field will be left unset and omitted during serialization. Servers older than v1.15 do not set this field. The intended use of the remainingItemCount is *estimating* the size of a collection. Clients should not rely on the remainingItemCount to be set or to be exact.
	RemainingItemCount *int64 `json:"remainingItemCount,omitempty"`
}

// ObjectMeta ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create.
type ObjectMeta struct {
	CreationTimestamp *string `json:"creationTimestamp,omitempty"`
	DeletionTimestamp *string `json:"deletionTimestamp,omitempty"`

	// Id id of the object
	Id *string `json:"id,omitempty"`

	// Labels Map of string keys and values that can be used to organize and categorize (scope and select) objects.
	Labels *map[string]string `json:"labels,omitempty"`
}

// Status Status is a return value for calls that don't return other objects.
type Status struct {
	// Message A human-readable description of the status of this operation.
	Message *string `json:"message,omitempty"`

	// Reason A machine-readable description of why this operation is in the "Failure" status. If this value is empty there is no information available. A Reason clarifies an HTTP status code but does not override it.
	Reason *string `json:"reason,omitempty"`

	// Status Status of the operation. One of: "Success" or "Failure". More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Status *string `json:"status,omitempty"`
}

// ListDevicesParams defines parameters for ListDevices.
type ListDevicesParams struct {
	// Continue An optional parameter to query more results from the server. The value of the paramter must match the value of the 'continue' field in the previous list response.
	Continue *string `form:"continue,omitempty" json:"continue,omitempty"`

	// LabelSelector A selector to restrict the list of returned objects by their labels. Defaults to everything.
	LabelSelector *string `form:"labelSelector,omitempty" json:"labelSelector,omitempty"`

	// Limit The maximum number of results returned in the list response. The server will set the 'continue' field in the list response if more results exist. The continue value may then be specified as parameter in a subesquent query.
	Limit *int32 `form:"limit,omitempty" json:"limit,omitempty"`
}