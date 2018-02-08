package murder

// Murder :
// Orchestra for queueing systems
// Provides availability via introducing multiple queues, locking and clearing
type Murder struct {
	crow          Crow
	queueSize     int
	lockTTL       int
	workerGroupID string
	queueAge      int
}

// Add :
// Create a job in any queue
func (m *Murder) Add(obj interface{}) {
	size := m.crow.QueueSize(m.workerGroupID)
	ageConfigured := m.AgeConfigured()
	var age int
	if (ageConfigured) {
		age = m.crow.QueueTimeSinceCreation(m.workerGroupID)
	}
	if size >= m.queueSize || (ageConfigured && age > m.queueAge) {
		queueName := newUUID()
		m.crow.MoveToReady(m.workerGroupID, queueName)
	}
	m.crow.AddToQueue(m.workerGroupID,  obj, ageConfigured)
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
		for {
			err := m.crow.ClearQueue(q, m.workerGroupID)
			if err != nil {
				continue
			}
			m.crow.RemoveLockKey(lockKey) // for cleaning up
			break
		}
	}
}

// Unlock :
// Unlock a queue, but not marking it as done
// Useful when a worker knows it is being killed and won't be able to finish the job
func (m *Murder) Unlock(lockKey string) {
	m.crow.RemoveLockKey(lockKey)
}

func (m *Murder) AgeConfigured() bool {
	return (m.queueAge > 0)
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


// MurderWithAge:
// Returns a new instance of murder with extra option of queue age

func NewMurderWithAge(bulkSize, TTL int, crow Crow, groupID string, queueAge int) *Murder {
	murder := NewMurder(bulkSize, TTL, crow, groupID)
	murder.queueAge = queueAge
	return murder
}
