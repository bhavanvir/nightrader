package main

// func main() {
// 	// This is a placeholder for the main function
// }

import (
	"fmt"
	"sync"
)

// Queue represents a basic queue structure.
type Queue struct {
	items []int
	lock  sync.RWMutex
}

// Enqueue adds an item to the end of the queue.
func (q *Queue) Enqueue(item int) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.items = append(q.items, item)
}

// Dequeue removes and returns the item from the front of the queue.
func (q *Queue) Dequeue() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.items) == 0 {
		return -1 // or any value indicating an empty queue
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

// UpdateItem updates a specific item in the queue.
func (q *Queue) UpdateItem(oldValue, newValue int) {
	q.lock.Lock()
	defer q.lock.Unlock()
	for i, item := range q.items {
		if item == oldValue {
			q.items[i] = newValue
			break
		}
	}
}

// PrintQueue prints the current items in the queue.
func (q *Queue) PrintQueue() {
	q.lock.RLock()
	defer q.lock.RUnlock()
	fmt.Println("Queue:", q.items)
}

func main() {
	myQueue := Queue{}

	// Enqueue some items
	myQueue.Enqueue(10)
	myQueue.Enqueue(20)
	myQueue.Enqueue(30)

	// Print the initial queue
	myQueue.PrintQueue()

	myQueue.Enqueue(35)
	myQueue.Enqueue(30)

	// Update an item in the queue
	myQueue.UpdateItem(20, 25)

	myQueue.Enqueue(40)

	// Print the updated queue
	myQueue.PrintQueue()
}