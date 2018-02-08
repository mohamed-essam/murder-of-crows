package murder

// Crow :
// Interface for any storage system for the orchestrator to use
type Crow interface {
	QueueSize(string) int                   // Query main queue size
	QueueTimeSinceCreation(string) int	// Query time since queue creation
	AddToQueue(*Murder, interface{})        // Add object to queue
	GetQueueContents(string) []string       // Retrieve all contents of queue
	ClearQueue(string, string) error        // Clear all queue contents
	CreateLockKey(string, string, int) bool // Create lock key for a queue, confirm if lock acquired, and set TTL
	IsLocked(string) bool                   // Check if a queue is locked
	FindQueueByKey(string) (string, bool)   // Get the queue for a lock key if exists
	ExtendLockKey(string, int)              // Extend TTL of lock key to value provided
	RemoveLockKey(string)                   // Removes a lock key if exists
	MoveToReady(string, string)             // Move a queue to ready to process queues
	GetReadyQueues(string) []string         // Get full queues
}
