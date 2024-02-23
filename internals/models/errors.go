package models

import (
    "errors"
)

// ErrNoRecord simplifies returning a specific error message when no matching
// database model is able to be retrieved.
var ErrNoRecord = errors.New("models: no matching record found")