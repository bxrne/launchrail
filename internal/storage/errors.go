package storage

import "errors"

// ErrRecordNotFound is returned when a requested record hash does not exist.
var ErrRecordNotFound = errors.New("record not found")
