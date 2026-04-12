package pool

import "sync"

// Resetter describes types that can reset their state.
type Resetter interface {
	Reset()
}

// Pool is a type-safe wrapper around sync.Pool for objects with Reset() method.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New creates a new Pool. The newFunc is called when the pool is empty and a new object is needed.
func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
	}
}

// Get returns an object from the pool.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put resets the object and returns it to the pool.
func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
