package api

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestApi(t *testing.T) {
	t.Run("test one publish : ten subscriptions", func(t *testing.T) {
		sb := NewSubPub()
		subscribers := make([]Subscription, 10)
		outputArray := make([]int, 0)
		wg := new(sync.WaitGroup)
		mu := new(sync.Mutex)
		for i := range subscribers {
			var err error
			wg.Add(1)
			subscribers[i], err = sb.Subscribe("1", func(msg any) {
				mu.Lock()
				outputArray = append(outputArray, i)
				mu.Unlock()
				defer wg.Done()
			})
			assert.NoError(t, err)
		}

		err := sb.Publish("1", 2)
		assert.NoError(t, err)
		wg.Wait()

		resArr := make([]int, len(subscribers))
		assert.Equal(t, len(resArr), len(outputArray))
	})

	t.Run("test ten publish : one subscription", func(t *testing.T) {
		sb := NewSubPub()
		outputArray := make([]int, 0)
		wg := new(sync.WaitGroup)
		mu := new(sync.Mutex)
		wg.Add(10)
		_, err := sb.Subscribe("1", func(msg any) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			outputArray = append(outputArray, msg.(int))
		})
		assert.NoError(t, err)

		for i := range 10 {
			err = sb.Publish("1", i)
			assert.NoError(t, err)
		}
		wg.Wait()

		resArr := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		assert.Equal(t, len(resArr), len(outputArray))
		assert.Equal(t, resArr, outputArray)
	})

	t.Run("test ten publish : two subscriptions with slow one", func(t *testing.T) {
		sb := NewSubPub()
		first := make([]int, 0)
		second := make([]int, 0)
		wg := new(sync.WaitGroup)
		mu := new(sync.Mutex)
		wg.Add(10)
		_, err := sb.Subscribe("1", func(msg any) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			first = append(first, msg.(int))
		})
		assert.NoError(t, err)
		_, err = sb.Subscribe("1", func(msg any) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			time.Sleep(1 * time.Second)
			second = append(second, msg.(int))
		})
		assert.NoError(t, err)

		for i := range 5 {
			err = sb.Publish("1", i)
			assert.NoError(t, err)
		}
		wg.Wait()

		resArr := []int{0, 1, 2, 3, 4}
		assert.Equal(t, resArr, first, "wrong fast subscriber")
		assert.Equal(t, resArr, second, "wrong slow subscriber")
	})
}

func TestClosing(t *testing.T) {
	t.Run("ok unsubscribe", func(t *testing.T) {
		sp := NewSubPub()
		s, err := sp.Subscribe("1", func(msg any) {
			fmt.Println(msg)
		})
		assert.NoError(t, err)
		s.Unsubscribe()
		err = sp.Publish("1", 1)
		assert.Error(t, err, "should return error")
	})

	t.Run("ok close", func(t *testing.T) {
		sp := NewSubPub()
		_, err := sp.Subscribe("1", func(msg any) {
			fmt.Println(msg)
		})
		assert.NoError(t, err)
		err = sp.Close(context.Background())
		assert.NoError(t, err)
	})

	t.Run("context cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		sp := NewSubPub()
		err := sp.Close(ctx)
		assert.ErrorIs(t, err, ctx.Err(), "should return deadline error")
		out := make([]int, 0)
		wg := new(sync.WaitGroup)
		mu := new(sync.Mutex)
		wg.Add(1)
		_, err = sp.Subscribe("1", func(msg any) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			out = append(out, msg.(int))
		})
		assert.NoError(t, err)
		err = sp.Publish("1", 1)
		wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, []int{1}, out)

	})
}
