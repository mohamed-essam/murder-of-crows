package murder

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gopkg.in/redis.v5"
)

type RedisCrow struct {
	Redis *redis.Client
}

func (c *RedisCrow) QueueSize(groupName string) int {
	size, _ := c.Redis.LLen(c.CurrentQueue(groupName)).Result()
	return int(size)
}

func (c *RedisCrow) CurrentQueue(groupName string) string {
	return fmt.Sprintf("murder::%s::mainQueue", groupName)
}

func (c *RedisCrow) AddToQueue(groupName string, obj interface{}) {
	marshalled, _ := json.Marshal(obj)
	c.Redis.LPush(c.CurrentQueue(groupName), marshalled).Result()
}

func (c *RedisCrow) GetQueueContents(queueName string) []string {
	contents, _ := c.Redis.LRange(fmt.Sprintf("murder::crows::%s", queueName), 0, -1).Result()
	return contents
}

func (c *RedisCrow) ClearQueue(queueName string, groupID string) error {
	_, err := c.Redis.Del(fmt.Sprintf("murder::crows::%s", queueName)).Result()
	if err == nil {
		c.Redis.SRem(fmt.Sprintf("murder::%s::ready", groupID), queueName).Result()
	} else {
		log.Printf("Error: %s", err.Error())
	}
	return err
}

func (c *RedisCrow) CreateLockKey(queueName string, lockKey string, TTL int) bool {
	locked, _ := c.Redis.SetNX(fmt.Sprintf("murder::crows::%s::locked", queueName), 1, time.Second*time.Duration(TTL)).Result()
	if locked {
		c.Redis.Set(fmt.Sprintf("murder::crows::%s::key", lockKey), queueName, time.Duration(0))
	}
	return locked
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
	queue, ok := c.FindQueueByKey(lockKey)
	c.Redis.Del(fmt.Sprintf("murder::crows::%s::key", lockKey))
	if ok {
		c.Redis.Del(fmt.Sprintf("murder::crows::%s::locked", queue))
	}
}

func (c *RedisCrow) MoveToReady(groupName, newName string) {
	ok, _ := c.Redis.SetNX(fmt.Sprintf("%s::lock", c.CurrentQueue(groupName)), "1", time.Duration(1)*time.Second).Result()
	if ok {
		c.Redis.Rename(c.CurrentQueue(groupName), fmt.Sprintf("murder::crows::%s", newName))
		c.Redis.SAdd(fmt.Sprintf("murder::%s::ready", groupName), newName)
		c.Redis.Del(fmt.Sprintf("%s::lock", c.CurrentQueue(groupName)))
	}
}

func (c *RedisCrow) GetReadyQueues(groupID string) []string {
	ret, _ := c.Redis.SMembers(fmt.Sprintf("murder::%s::ready", groupID)).Result()
	return ret
}
