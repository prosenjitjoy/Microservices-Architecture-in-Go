package repository

import "errors"

// ErrorNotFound is retured when a requested record is not found.
var ErrNotFound = errors.New("not found")
