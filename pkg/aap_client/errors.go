package aap_client

import "errors"

var (
	ErrNotFound  = errors.New("resource not found")
	ErrForbidden = errors.New("resource forbidden")
)
