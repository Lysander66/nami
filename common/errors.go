package common

import "errors"

var (
	// ErrMaxDepth is the error type for exceeding max depth
	ErrMaxDepth = errors.New("max depth limit reached")
	// ErrEmptyProxyURL is the error type for empty Proxy URL list
	ErrEmptyProxyURL = errors.New("proxy URL list is empty")
)
