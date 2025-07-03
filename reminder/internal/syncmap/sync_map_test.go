package syncmap_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/internal/syncmap"
	"github.com/stretchr/testify/assert"
)

func TestBasicOperations(t *testing.T) {
	t.Run("Set and Get existing key", func(t *testing.T) {
		m := syncmap.New[string, int]()
		m.Set("key1", 1)
		val, ok := m.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 1, val)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		m := syncmap.New[string, int]()
		_, ok := m.Get("non-existent")
		assert.False(t, ok)
	})

	t.Run("Delete existing key", func(t *testing.T) {
		m := syncmap.New[string, int]()
		m.Set("key1", 1)
		m.Delete("key1")
		_, ok := m.Get("key1")
		assert.False(t, ok)
	})
}

func TestRange(t *testing.T) {
	t.Run("iterate all elements", func(t *testing.T) {
		m := syncmap.New[string, int]()
		testData := map[string]int{
			"key1": 1,
			"key2": 2,
			"key3": 3,
		}

		for k, v := range testData {
			m.Set(k, v)
		}

		visited := make(map[string]int)
		m.Range(func(key string, val int) bool {
			visited[key] = val
			return true
		})

		assert.Equal(t, testData, visited)

		m.Delete("key1")
		m.Delete("key2")
		m.Delete("key3")
		m.Range(func(key string, val int) bool {
			t.Fatal("keys should not exist")
			return true
		})
	})

	t.Run("early exit", func(t *testing.T) {
		m := syncmap.New[string, int]()
		for i := range 3 {
			m.Set(fmt.Sprintf("key%d", i), i)
		}

		count := 0
		m.Range(func(key string, val int) bool {
			count++
			return false
		})
		assert.Equal(t, 1, count)
	})
}

func TestApply(t *testing.T) {
	t.Run("apply to existing key", func(t *testing.T) {
		m := syncmap.New[string, int]()
		m.Set("key1", 1)

		m.Apply("key1", func(val int) int {
			return val * 2
		})

		val, ok := m.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 2, val)
	})

	t.Run("apply to non-existent key", func(t *testing.T) {
		m := syncmap.New[string, int]()

		m.Apply("non-existent", func(val int) int {
			return val * 2
		})

		_, ok := m.Get("non-existent")
		assert.False(t, ok)
	})
}

func TestConcurrentAccess(t *testing.T) {
	const n = 1000

	t.Run("concurrent writes to different keys", func(t *testing.T) {
		m := syncmap.New[int, int]()
		var wg sync.WaitGroup

		for i := range n {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				m.Set(i, i)
			}(i)
		}
		wg.Wait()

		count := 0
		m.Range(func(key, val int) bool {
			count++
			assert.Equal(t, key, val)
			return true
		})
		assert.Equal(t, n, count)
	})

	t.Run("concurrent reads and writes to different keys", func(t *testing.T) {
		m := syncmap.New[int, int]()
		var wg sync.WaitGroup

		// First set initial values
		for i := range n {
			m.Set(i, i)
		}

		// Then test concurrent reads and writes
		for i := range n {
			wg.Add(2)
			go func(i int) {
				defer wg.Done()
				m.Set(i, i*2)
			}(i)
			go func(i int) {
				defer wg.Done()
				val, ok := m.Get(i)
				assert.True(t, ok)
				assert.True(t, val == i || val == i*2)
			}(i)
		}
		wg.Wait()

		// Verify final state
		for i := range n {
			val, ok := m.Get(i)
			assert.True(t, ok)
			assert.Equal(t, i*2, val, "Final value for key %d should be %d", i, i*2)
		}
	})

	t.Run("concurrent delete operations", func(t *testing.T) {
		m := syncmap.New[string, int]()
		var wg sync.WaitGroup

		// Initialize map with values
		for i := range n {
			m.Set(fmt.Sprintf("key%d", i), i)
		}

		// Start goroutines that will concurrently delete and check values
		for i := range n {
			wg.Add(2)
			go func(i int) {
				defer wg.Done()
				m.Delete(fmt.Sprintf("key%d", i))
			}(i)
			go func(i int) {
				defer wg.Done()
				key := fmt.Sprintf("key%d", i)
				_, ok := m.Get(key)
				// We can't assert anything about ok here because we don't know
				// if Delete has already happened or not
				_ = ok
			}(i)
		}
		wg.Wait()

		// After all operations complete, verify that all keys were deleted
		deletedCount := 0
		m.Range(func(key string, val int) bool {
			deletedCount++
			return true
		})
		assert.Zero(t, deletedCount, "All keys should have been deleted")
	})
}

func TestConcurrentAccessToSameKey(t *testing.T) {
	const n = 1000
	const key = "test-key"

	t.Run("concurrent increments", func(t *testing.T) {
		m := syncmap.New[string, int]()
		var wg sync.WaitGroup

		m.Set(key, 0)

		for range n {
			wg.Add(1)
			go func() {
				defer wg.Done()
				m.Apply(key, func(val int) int {
					return val + 1
				})
			}()
		}
		wg.Wait()

		val, ok := m.Get(key)
		assert.True(t, ok)
		assert.Equal(t, n, val, "After %d concurrent increments, value should be %d, got %d", n, n, val)
	})

	t.Run("concurrent set operations", func(t *testing.T) {
		m := syncmap.New[string, int]()
		var wg sync.WaitGroup
		resultChan := make(chan int, n)

		// Start goroutines that will concurrently write and read values
		for i := range n {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				m.Set(key, i)

				// Get the value after setting it
				val, ok := m.Get(key)
				assert.True(t, ok)

				select {
				case resultChan <- val:
				default:
					t.Errorf("Failed to send value to channel: %d", val)
				}
			}(i)
		}

		// Wait for all operations to complete
		wg.Wait()

		close(resultChan)

		// Verify all read values were valid
		valuesCount := 0
		for val := range resultChan {
			valuesCount++
			assert.GreaterOrEqual(t, val, 0)
			assert.Less(t, val, n)
		}
		assert.Equal(t, n, valuesCount, "Expected to receive exactly %d values", n)
	})

	t.Run("simultaneous get and set operations", func(t *testing.T) {
		m := syncmap.New[string, int]()
		var setWg, getWg sync.WaitGroup

		m.Set(key, 0)

		// Launch set operations
		for i := range n {
			setWg.Add(1)
			go func(i int) {
				defer setWg.Done()
				m.Set(key, i)
			}(i)
		}

		// Launch get operations
		for range n {
			getWg.Add(1)
			go func() {
				defer getWg.Done()
				val, ok := m.Get(key)
				assert.True(t, ok)
				assert.GreaterOrEqual(t, val, 0)
				assert.Less(t, val, n)
			}()
		}

		setWg.Wait()
		getWg.Wait()
	})
}
