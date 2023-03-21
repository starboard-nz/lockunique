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

func Benchmark1point6mill(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.Run("1.6 million locks", func(b *testing.B) {
			// test taking 1.6 million locks
			l := lockunique.NewLockUnique[int64]()
			totalLocks := 1600000
			//	Create a map of 1.6 million vessels
			for i := 0; i < totalLocks; i++ {
				// lock all of them
				l.Lock(int64(i))
			}

			// now unlock all of them
			for i := 0; i < totalLocks; i++ {
				err := l.Unlock(int64(i))
				if err != nil {
					b.Errorf("err = %v", err)
				}
			}
		})
		b.Run("1.6 million locks unlocks (map version)", func(b *testing.B) {
			// test taking 1.6 million locks
			lockMap := make(map[int64]*sync.Mutex)
			totalLocks := 1600000
			//	Create a map of 1.6 million vessels
			for i := 0; i < totalLocks; i++ {
				// create the lock
				lockMap[int64(i)] = &sync.Mutex{}
				lockMap[int64(i)].Lock()
			}

			// now unlock all of them
			for i := 0; i < totalLocks; i++ {
				lockMap[int64(i)].Unlock()
			}
		})

	}
}

func BenchmarkBasic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := lockunique.NewLockUnique[int32]()

		const maxID = 1000

		var n [maxID]int32

		wg := &sync.WaitGroup{}

		for i := 0; i < 100000; i++ {
			wg.Add(1)
			vID := rand.Int31n(maxID-1) + 1

			if i%1000 == 0 {
				// makes it keep switching back to an array as the queue clears
				time.Sleep(50 * time.Millisecond)
			}

			go func(vID int32) {
				l.Lock(vID)
				n0 := atomic.AddInt32(&(n[vID-1]), 1)
				if n0 != 1 {
					b.Errorf("n0 = %d", n0)
				}
				time.Sleep(2 * time.Millisecond)
				n0 = atomic.AddInt32(&(n[vID-1]), -1)
				if n0 != 0 {
					b.Errorf("n0 = %d", n0)
				}

				l.Unlock(vID)
				wg.Done()
			}(vID)
		}

		wg.Wait()
	}
}

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
			//time.Sleep(2 * time.Millisecond)
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

func TestLockUnlockSameID(t *testing.T) {
	l := lockunique.NewLockUnique[int32]()
	id := int32(123)

	var (
		locking, unlocking int32
	)

	count := 100000
	wg := &sync.WaitGroup{}
	queue := make(chan struct{}, 10)

	i := 0
	for ; i < count; i++ {
		queue <-struct{}{}
		wg.Add(1)

		go func() {
			l.Lock(id)
			atomic.AddInt32(&locking, 1)

			l.Unlock(id)
			atomic.AddInt32(&unlocking, 1)

			<-queue

			wg.Done()
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <-struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(10*time.Second):
		t.Logf("i = %d; locking = %d; unlocking = %d", i, locking, unlocking)
		t.Errorf("Timed out.")
	}

	if locking != int32(count) {
		t.Errorf("incorrect result")
	}

	if unlocking != int32(count) {
		t.Errorf("incorrect result")
	}
}
