package databases

import "errors"

var (
	/* those errors occurs only when either
	of them is empty and the other is populated
	*/
	ErrEmptydbKey   = errors.New("key is Empty")
	ErrEmptydbValue = errors.New("value is Empty")
	ErrInvalidDB    = errors.New("invalid DB")
)
