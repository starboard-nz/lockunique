package lockunique_test

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/starboard-nz/lockunique"
)

func TestMain(m *testing.M) {
	go RunDebugServer(8888)

	m.Run()
}

func TestBasic(t *testing.T) {
	l := lockunique.NewLockUnique[int32]()

	l.Lock(int32(1))

	done := make(chan struct{})

	go func() {
		l.Lock(int32(1))

		done <- struct{}{}
	}()

	time.Sleep(10 * time.Millisecond)

	select {
	case <-done:
		t.Errorf("Same ID locked twice")
	default:
	}

	l.Lock(int32(2))
	
	err := l.Unlock(int32(2))
	assert.Nil(t, err)

	select {
	case <-done:
		t.Errorf("Same ID locked twice")
	default:
	}

	err = l.Unlock(int32(2))
	assert.NotNil(t, err)

	err = l.Unlock(int32(1))
	assert.Nil(t, err)

	time.Sleep(10 * time.Millisecond)

	select {
	case <-done:
	default:
		t.Errorf("Lock should have succeeded")
	}

	err = l.Unlock(int32(1))
	assert.Nil(t, err)

	err = l.Unlock(int32(1))
	assert.NotNil(t, err)
}

func TestLockUnlock(t *testing.T) {
	l := lockunique.NewLockUnique[int32]()

	const maxID = 1000

	var n [maxID]int32

	wg := &sync.WaitGroup{}

	for i := 0; i < 100000; i++ {
		wg.Add(1)
		vID := rand.Int31n(maxID-1) + 1

		if i % 1000 == 0 {
			// makes it keep switching back to an array as the queue clears
			time.Sleep(50 * time.Millisecond)
		}

		go func(vID int32) {
			l.Lock(vID)
			n0 := atomic.AddInt32(&(n[vID-1]), 1)
			if n0 != 1 {
				t.Errorf("n0 = %d", n0)
			}
			time.Sleep(2 * time.Millisecond)
			n0 = atomic.AddInt32(&(n[vID-1]), -1)
			if n0 != 0 {
				t.Errorf("n0 = %d", n0)
			}

			l.Unlock(vID)
			wg.Done()
		}(vID)
	}

	wg.Wait()
}

func TestLockUnlock2(t *testing.T) {
	l := lockunique.NewLockUnique[int32]()

	const maxID = 1000

	var n [maxID]int32

	wg := &sync.WaitGroup{}

	for i := 0; i < 100000; i++ {
		wg.Add(1)
		vID := rand.Int31n(maxID-1) + 1

		go func(vID int32) {
			l.Lock(vID)
			n0 := atomic.AddInt32(&(n[vID-1]), 1)
			if n0 != 1 {
				t.Errorf("n0 = %d", n0)
			}
			time.Sleep(2 * time.Millisecond)
			n0 = atomic.AddInt32(&(n[vID-1]), -1)
			if n0 != 0 {
				t.Errorf("n0 = %d", n0)
			}

			l.Unlock(vID)
			wg.Done()
		}(vID)
	}

	wg.Wait()
}
