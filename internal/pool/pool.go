package pool

import "sync"

// Resetter – интерфейс, который должен реализовывать тип, хранимый в пуле.
type Resetter interface {
	Reset()
}

// Pool – generic-пул для переиспользования объектов.
// Обеспечивает автоматический сброс объекта перед помещением обратно в пул.
type Pool[T Resetter] struct {
	p sync.Pool
}

// New создаёт Pool с конструктором для создания новых объектов, когда пул пуст.
func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

// Get извлекает объект из пула (или создаёт новый конструктором, если пул пуст).
func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

// Put сбрасывает объект вызовом Reset() и помещает его обратно в пул.
func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.p.Put(x)
}
