# Murder of crows
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/670a0bac960d47a1bdd6c4902955ceed)](https://app.codacy.com/app/messam/murder-of-crows?utm_source=github.com&utm_medium=referral&utm_content=mohamed-essam/murder-of-crows&utm_campaign=badger)

Miniature extendable queue orchestration library, capable of creating multiple equal sized queues that can be consumed in parallel without needing to lock the entire queue for processing from a single worker.

## Usage

Create a Murder object to manage queues, assigning a Crow object as an adapter to communicate with any atomic-enabled system, a RedisCrow is built into the library.

```go
import "github.com/mohamed-essam/murder-of-crows"
import "redis.v5"
import "fmt"

func main() {
  redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", Password: "", DB: 0})
  redisCrow := &murder.RedisCrow{Redis: redisClient}
  // Create a new murder instance with 100 as a bulk size, 1 second TTL and a worker_group_name
  // Worker group name separates different queue systems by prefixing them
  // 100 is a rough number as no locking is used in enqueueing to speed this operation up, the actual queue size will be >= 100
  murderInstance := murder.NewMurder(100, 1, redisCrow, "worker_group_name") 

  for (i := 0; i < 100; i++) {
    murderInstance.Add(i)
  }

  lockKey, locked := murderInstance.Lock() // This locks a full queue giving only the worker with lockKey access to it

  if (locked) { // this may be false if there are no full queues
    jobs := murderInstance.Get(lockKey) // jobs are always returned as an array of strings

    for _, obj := range jobs { // Prints out 0 through 99
      fmt.Printf("%s\n", obj)
    }
  }
}
```
