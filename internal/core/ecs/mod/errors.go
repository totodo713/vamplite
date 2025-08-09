package mod

import (
	"errors"
	"fmt"
)

var (
	// Entity関連エラー
	ErrEntityLimitExceeded    = errors.New("entity limit exceeded")
	ErrEntityPermissionDenied = errors.New("entity permission denied")
	ErrSystemEntityAccess     = errors.New("system entity access denied")

	// Component関連エラー
	ErrComponentNotAllowed       = errors.New("component not allowed")
	ErrComponentPermissionDenied = errors.New("component permission denied")

	// Query関連エラー
	ErrQueryLimitExceeded   = errors.New("query limit exceeded")
	ErrQueryTimeoutExceeded = errors.New("query timeout exceeded")

	// System関連エラー
	ErrSystemExecutionTimeExceeded = errors.New("execution time exceeds limit")
	ErrSystemMemoryLimitExceeded   = errors.New("system memory limit exceeded")
	ErrSecurityViolation           = errors.New("security violation")

	// MOD関連エラー
	ErrMemoryLimitExceeded = errors.New("memory limit exceeded")
	ErrModNotFound         = errors.New("mod not found")
	ErrModAlreadyExists    = errors.New("mod already exists")
)

// SecurityError セキュリティ違反エラー
type SecurityError struct {
	ModID     string
	Operation string
	Reason    string
}

func (e *SecurityError) Error() string {
	return fmt.Sprintf("security violation in mod %s: %s (%s)", e.ModID, e.Operation, e.Reason)
}

// ResourceError リソース制限エラー
type ResourceError struct {
	ModID    string
	Resource string
	Current  int64
	Limit    int64
}

func (e *ResourceError) Error() string {
	return fmt.Sprintf("resource limit exceeded in mod %s: %s (%d/%d)", e.ModID, e.Resource, e.Current, e.Limit)
}
