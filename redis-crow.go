package murder

import (
	"fmt"
	"time"

	"gopkg.in/redis.v5"
)

type RedisCrow struct {
	Redis *redis.Client
}

func (c *RedisCrow) CreateQueue(queueName string) error {
	_, err := c.Redis.LPush("murder::crows", queueName).Result()
	return err
}

func (c *RedisCrow) QueueSize(queueName string) int {
	size, _ := c.Redis.LLen(fmt.Sprintf("murder::crows::%s", queueName)).Result()
	return int(size)
}

func (c *RedisCrow) AddToQueue(queueName string, obj interface{}) {
	marshalled, _ := json.Marshal(obj)
	c.Redis.LPush(fmt.Sprintf("murder::crows::%s", queueName), marshalled).Result()
}

func (c *RedisCrow) GetQueueContents(queueName string) []string {
	contents, _ := c.Redis.LRange(fmt.Sprintf("murder::crows::%s", queueName), 0, -1).Result()
	return contents
}

func (c *RedisCrow) ClearQueue(queueName string) error {
	_, err := c.Redis.Del(fmt.Sprintf("murder::crows::%s", queueName)).Result()
	c.Redis.LRem("murder::crows", 1, queueName).Result()
	return err
}

func (c *RedisCrow) CreateLockKey(queueName string, lockKey string, TTL int) bool {
	locked, _ := c.Redis.SetNX(fmt.Sprintf("murder::crows::%s::locked", queueName), 1, time.Second*time.Duration(TTL)).Result()
	if locked {
		c.Redis.Set(fmt.Sprintf("murder::crows::%s::key", lockKey), queueName, time.Duration(0))
	}
	return locked
}

func (c *RedisCrow) GetQueues() []string {
	queues, _ := c.Redis.LRange("murder::crows", 0, -1).Result()
	return queues
}

func (c *RedisCrow) IsLocked(queueName string) bool {
	locked, err := c.Redis.Get(fmt.Sprintf("murder::crows::%s::locked", queueName)).Result()
	if locked == "1" && err == nil {
		return true
	}
	return false
}

func (c *RedisCrow) FindQueueByKey(lockKey string) (string, bool) {
	queue, err := c.Redis.Get(fmt.Sprintf("murder::crows::%s::key", lockKey)).Result()
	return queue, err == nil
}

func (c *RedisCrow) ExtendLockKey(lockKey string, TTL int) {
	queue, ok := c.FindQueueByKey(lockKey)
	if ok {
		c.Redis.Expire(fmt.Sprintf("murder::crows::%s::locked", queue), time.Second*time.Duration(TTL))
	}
}

func (c *RedisCrow) RemoveLockKey(lockKey string) {
	c.Redis.Del(fmt.Sprintf("murder::crows::%s::key", lockKey))
	queue, ok := c.FindQueueByKey(lockKey)
	if ok {
		c.Redis.Del(fmt.Sprintf("murder::crows::%s::locked", queue))
	}
}
