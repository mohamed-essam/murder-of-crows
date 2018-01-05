package murder

// Murder :
// Orchestra for queueing systems
// Provides availability via introducing multiple queues, locking and clearing
type Murder struct {
	crow          Crow
	queueSize     int
	lockTTL       int
	workerGroupID string
}

// Add :
// Create a job in any queue
func (m *Murder) Add(obj interface{}) {
	queues := m.crow.GetQueues(m.workerGroupID)
	if len(queues) > 0 {
		q := queues[len(queues)-1]
		size := m.crow.QueueSize(q)
		if !m.crow.IsLocked(q) && size < m.queueSize { // Queue is unlocked and can be added to
			m.crow.AddToQueue(q, obj)
			if size+1 >= m.queueSize {
				m.crow.MoveToReady(q, m.workerGroupID)
			}
			return
		}
	}
	// No suitable queues found, create a new queue and add to it
	queueName := newUUID()
	m.crow.CreateQueue(queueName, m.workerGroupID)
	m.crow.AddToQueue(queueName, obj)
}

// Lock :
// Lock a queue returning a lock key that is needed for acknowledging the processing of the queue
// If no queue is ready to process, returns empty string and false
func (m *Murder) Lock() (string, bool) {
	queues := m.crow.GetReadyQueues(m.workerGroupID)
	for _, q := range queues {
		if !m.crow.IsLocked(q) { // Queue is unlocked and can be processed
			lockKey := newUUID()
			ok := m.crow.CreateLockKey(q, lockKey, m.lockTTL)
			if ok {
				return lockKey, true
			}
		}
	}
	return "", false
}

// Get :
// Get contents of a queue given its lock key
// Ensuring the worker locked the queue and acquired the lock key
func (m *Murder) Get(lockKey string) []string {
	q, ok := m.crow.FindQueueByKey(lockKey)
	if ok {
		return m.crow.GetQueueContents(q)
	}
	return []string{}
}

// Ack :
// Acknowledge processing of a queue lock extending its time to kill
// Useful for long running jobs
func (m *Murder) Ack(lockKey string) {
	m.crow.ExtendLockKey(lockKey, m.lockTTL)
}

// Mark :
// Mark a locked queue as done, and its jobs disposable
func (m *Murder) Mark(lockKey string) {
	q, ok := m.crow.FindQueueByKey(lockKey)
	if ok {
		m.crow.ClearQueue(q, m.workerGroupID)
		m.crow.RemoveLockKey(lockKey) // for cleaning up
	}
}

// Unlock :
// Unlock a queue, but not marking it as done
// Useful when a worker knows it is being killed and won't be able to finish the job
func (m *Murder) Unlock(lockKey string) {
	m.crow.RemoveLockKey(lockKey)
}

// NewMurder :
// Returns a new instance of murder with the given options
func NewMurder(bulkSize, TTL int, crow Crow, groupID string) *Murder {
	return &Murder{
		crow:          crow,
		queueSize:     bulkSize,
		lockTTL:       TTL,
		workerGroupID: groupID,
	}
}
