package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	mtx      *sync.Mutex
}

type cacheValue struct {
	value interface{}
	key   Key
}

func (lru *lruCache) Set(key Key, value interface{}) bool {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()
	if item, ok := lru.items[key]; ok {
		item.Value = &cacheValue{
			value: value,
			key:   key,
		} // refresh value
		lru.queue.MoveToFront(item)
		return true
	}
	chVal := &cacheValue{
		value: value,
		key:   key,
	}

	listItem := lru.queue.PushFront(chVal)
	lru.items[key] = listItem

	if lru.queue.Len() > lru.capacity {
		lastElem := lru.queue.Back() // must exists!!
		chVal = lastElem.Value.(*cacheValue)
		backKey := chVal.key
		lru.queue.Remove(lastElem) // remove the last
		delete(lru.items, backKey)
	}

	return false
}

func (lru *lruCache) Clear() {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()
	lru.items = make(map[Key]*ListItem, lru.capacity)
	lru.queue = NewList()
}

func (lru *lruCache) Get(key Key) (interface{}, bool) {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()
	if item, ok := lru.items[key]; ok {
		lru.queue.MoveToFront(item)
		chVal := item.Value.(*cacheValue)
		return chVal.value, true
	}
	return nil, false
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
		mtx:      &sync.Mutex{},
	}
}
