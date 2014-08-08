package r2dq

import (
	"errors"
	"fmt"

	"github.com/canthefason/irc-k/config"
	"gopkg.in/redis.v2"
)

const (
	WAITING_QUEUE    = "waitingQueue"
	PROCESSING_QUEUE = "processingQueue"
)

var (
	redisConn   *redis.Client
	ErrNotFound = errors.New("not found")
)

type Queue struct {
	Prefix string
}

func init() {
	MustInitRedis()
}

func MustInitRedis() {
	redisConf := config.Conf.Redis
	redisConn = redis.NewTCPClient(&redis.Options{
		Addr:     redisConf.Server + ":" + redisConf.Port,
		Password: "",
		DB:       int64(redisConf.DB),
	})
}

func MustGetRedis() *redis.Client {
	return redisConn
}

func (q *Queue) Queue(value string) error {
	res := redisConn.LPush(q.waitingQueueKey(), value)

	return res.Err()
}

func (q *Queue) Dequeue() (string, error) {
	res := redisConn.RPopLPush(q.waitingQueueKey(), q.procQueueKey())

	if res.Err() != nil && res.Err() != redis.Nil {
		return "", res.Err()
	}

	return res.Val(), nil
}

func (q *Queue) Ack(val string) error {
	res := redisConn.LRem(q.procQueueKey(), 1, val)

	if res.Err() != nil {
		if res.Err() == redis.Nil {
			return ErrNotFound
		}
		return res.Err()
	}

	return nil
}

func (q *Queue) Close() {
	q.gracefulShutdown()
	redisConn.Close()
}

func (q *Queue) gracefulShutdown() {
	res := redisConn.RPopLPush(q.procQueueKey(), q.waitingQueueKey())
	for res.Val() != "" {
		res = redisConn.RPopLPush(q.procQueueKey(), q.waitingQueueKey())
	}

	if res.Err() != redis.Nil {
		panic(res.Err())
	}
}

func (q *Queue) waitingQueueKey() string {
	return q.keyWithPrefix(WAITING_QUEUE)
}

func (q *Queue) procQueueKey() string {
	return q.keyWithPrefix(PROCESSING_QUEUE)
}

func (q *Queue) keyWithPrefix(queue string) string {
	return fmt.Sprintf("%s:%s", q.Prefix, queue)
}
