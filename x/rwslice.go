package x

import "sync"

// RWSlice is a type for a thread-safe Go slice. It tries to be short and simple.
// Tip: It's useful to create a type alias (it allows it):
//
//	type testersSlice = [int]testing.TB
//
// Which shortens and makes easier to read its usage:
//
//	x.TxSlice(testers, func(s testersSlice) {
//	    s[goid()] = val
//	})
type RWSlice[S ~[]T, T any] struct {
	sync.RWMutex
	s S
}

// NewRWSlice creates a new thread-safe slice that's as simple as possible. The
// first version had only two functions Tx and Rx to allow interact with the
// slice.
func NewRWSlice[S ~[]T, T any](size ...int) *RWSlice[S, T] {
	// build in make() have to deal by us
	switch len(size) {
	case 2:
		return &RWSlice[S, T]{s: make([]T, size[0], size[1])}
	default:
		return &RWSlice[S, T]{s: make([]T, size[0])}
	}
}

// TxSlice executes a critical section during the function given as an argument.
// This critical section allows the slice be updated. If you only need to read
// the slice please use the Rx function that's for a read-only critical section.
func TxSlice[S ~[]T, T any](s *RWSlice[S, T], f func(s S)) {
	s.Lock()
	defer s.Unlock()
	f(s.s)
}

// AddSlice sets a key value pair to the slice.
func AddSlice[S ~[]T, T any](s *RWSlice[S, T], val T) T {
	s.Lock()
	defer s.Unlock()
	s.s = append(s.s, val)
	return val
}

// SetSlice sets a key value pair to the slice.
func SetSlice[S ~[]T, T any](s *RWSlice[S, T], key int, val T) T {
	s.Lock()
	defer s.Unlock()
	s.s[key] = val
	return val
}

// RxSlice executes thread-safe read transaction i.e. critical section for the
// RWSlice.
func RxSlice[S ~[]T, T any](s *RWSlice[S, T], f func(s S)) {
	s.RLock()
	defer s.RUnlock()
	f(s.s)
}

// GetSlice reads thread-safely value from index key.
func GetSlice[S ~[]T, T any](s *RWSlice[S, T], key int) T {
	s.RLock()
	defer s.RUnlock()
	return s.s[key]
}
