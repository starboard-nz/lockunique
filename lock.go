package lockunique

import (
	"errors"
	"fmt"
	"sync"
)

const maxArraySize = 50

var ErrNotLocked = errors.New("id not locked")

type lockID[T comparable] struct {
	id       T
	queue    int32
	unlocked chan struct{}
}

type LockUnique[T comparable] struct {
	lockArr  []*lockID[T]
	lockMap  map[T]*lockID[T]
	mu       sync.Mutex
	valid    bool
	useMap   bool
	nextFree int
}

// NewLockUnique creates a new LockUnique[T]. Using lock := &lockunique.LockUnique{} can be use as well.
func NewLockUnique[T comparable]() *LockUnique[T] {
	return &LockUnique[T]{
		lockArr: make([]*lockID[T], 0, maxArraySize),
		valid:   true,
	}
}

// Lock acquires a lock for the given id.
func (l *LockUnique[T]) Lock(id T) {
	l.mu.Lock()

	if !l.valid {
		l.lockArr = make([]*lockID[T], maxArraySize)
		l.valid = true
	}

	var lock *lockID[T]

	if l.useMap {
		lock = l.lockMap[id]
	} else {
		for i := 0; i < l.nextFree; i++ {
			if l.lockArr[i].id == id {
				lock = l.lockArr[i]

				break
			}
		}
	}

	if lock != nil {
		// id is locked already
		lock.queue++
		l.mu.Unlock()

		// wait for the current lock to be deleted
		<-lock.unlocked

		return
	}

	// if we are there, there is no current lock for this id
	lock = &lockID[T]{
		id:       id,
		unlocked: make(chan struct{}, 1), // cap of 1 so the unlocker never has to wait
	}

	if l.useMap {
		l.lockMap[id] = lock
	} else {
		// using an array
		if l.nextFree >= len(l.lockArr) {
			if l.nextFree < maxArraySize {
				l.lockArr = append(l.lockArr, lock)
			} else {
				// switch to using a map
				l.useMap = true
				l.lockMap = make(map[T]*lockID[T])

				for _, lptr := range l.lockArr {
					l.lockMap[lptr.id] = lptr
				}

				l.lockArr = l.lockArr[:0]
				l.lockMap[id] = lock
				l.nextFree = 0
			}
		} else {
			l.lockArr[l.nextFree] = lock
		}

		l.nextFree++
	}

	// take first place in the queue
	lock.queue = 1
	l.mu.Unlock()
}

// Unlock releases the lock for the given id. If the id is not currently locked then an lockunique.ErrNotLocked is returned.
func (l *LockUnique[T]) Unlock(id T) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.useMap {
		lock, found := l.lockMap[id]
		if !found {
			return fmt.Errorf("%w: id = %v", ErrNotLocked, id)
		}
		// remove this lock from the queue
		lock.queue--

		if lock.queue == 0 {
			delete(l.lockMap, id)

			if len(l.lockMap) < maxArraySize / 2 {
				// revert to array
				for _, lptr := range l.lockMap {
					l.lockArr = append(l.lockArr, lptr)
				}

				l.useMap = false
				l.nextFree = len(l.lockArr)
				l.lockMap = nil
			}

			return nil
		}

		// the next lock in the queue can proceed
		lock.unlocked <- struct{}{}

		return nil
	}

	// using an array
	for i := 0; i < l.nextFree; i++ {
		if l.lockArr[i].id == id {
			// remove this lock from the queue
			l.lockArr[i].queue--

			if l.lockArr[i].queue == 0 {
				// no other locks waiting for this ID, remove from locks
				if i == l.nextFree-1 {
					l.lockArr[i] = nil
				} else {
					l.lockArr[i] = l.lockArr[l.nextFree-1]
				}

				l.nextFree--

				return nil
			}

			// the next lock in the queue can proceed
			l.lockArr[i].unlocked <- struct{}{}

			return nil
		}
	}

	return fmt.Errorf("%w: id = %v", ErrNotLocked, id)
}
