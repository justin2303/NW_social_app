package workerpool

import (
	"sync"
)

type WorkerPool struct {
	workerCount int
	tasks       chan func() // Channel to hold tasks (functions)
	wg          sync.WaitGroup
}

// NewWorkerPool initializes a new worker pool
func NewWorkerPool(workerCount int) *WorkerPool {
	pool := &WorkerPool{
		workerCount: workerCount,
		tasks:       make(chan func()), // Create a channel for tasks
	}

	// Start the workers
	for i := 0; i < workerCount; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

// worker processes tasks from the tasks channel
func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for task := range p.tasks {
		task() // Execute the task
	}
}

// Enqueue adds a task to the pool
func (p *WorkerPool) Enqueue(task func()) {
	p.tasks <- task
}

// Wait waits for all workers to finish
func (p *WorkerPool) Wait() {
	close(p.tasks) // Close the tasks channel to stop workers
	p.wg.Wait()    // Wait for all workers to finish
}
