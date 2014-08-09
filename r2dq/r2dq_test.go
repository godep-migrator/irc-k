package r2dq

import "testing"

func NewQueue() *Queue {
	q := new(Queue)
	q.Prefix = "test"

	return q
}

func tearDown(q *Queue) {
	redisConn.Del(q.waitingQueueKey())
	redisConn.Del(q.procQueueKey())
}

func TestQueue(t *testing.T) {
	q := NewQueue()
	defer tearDown(q)

	q.Queue("drteeth")

	res := redisConn.LLen(q.waitingQueueKey())
	if res.Val() != 1 {
		t.Errorf("Expected %d but got %d", 1, res.Val())
	}

	q.Queue("floyd")
	res = redisConn.LLen(q.waitingQueueKey())
	if res.Val() != 2 {
		t.Errorf("Expected %d but got %d", 2, res.Val())
	}
}

func TestDequeue(t *testing.T) {
	q := NewQueue()
	defer tearDown(q)

	q.Queue("drteeth")
	length := redisConn.LLen(q.waitingQueueKey())
	if length.Val() != 1 {
		t.Errorf("Expected %d but got %d", 1, length.Val())
	}

	q.Queue("floyd")
	length = redisConn.LLen(q.waitingQueueKey())
	if length.Val() != 2 {
		t.Errorf("Expected %d but got %d", 2, length.Val())
	}

	res, err := q.Dequeue()
	if err != nil {
		t.Errorf("Expected nil but got %s", err)
		t.FailNow()
	}

	if res != "drteeth" {
		t.Errorf("Expected %s but got %s", "drteeth", res)
	}

	length = redisConn.LLen(q.procQueueKey())
	if length.Val() != 1 {
		t.Errorf("Expected %d but got %d", 1, length.Val())
	}

	res, err = q.Dequeue()
	if err != nil {
		t.Errorf("Expected nil but got %s", err)
	}

	if res != "floyd" {
		t.Errorf("Expected %s but got %s", "floyd", res)
	}

	length = redisConn.LLen(q.procQueueKey())
	if length.Val() != 2 {
		t.Errorf("Expected %d but got %d", 2, length.Val())
	}

	res, err = q.Dequeue()
	if err != nil {
		t.Errorf("Expected nil but got %s", err)
	}

	if res != "" {
		t.Errorf("Expected empty queue, but got %s", res)
	}

}

func TestAck(t *testing.T) {
	q := NewQueue()
	defer tearDown(q)

	q.Queue("drteeth")
	q.Queue("floyd")

	q.Dequeue()
	q.Dequeue()

	err := q.Ack("floyd")
	if err != nil {
		t.Errorf("Expected nil but got %s", err)
	}

	length := redisConn.LLen(q.procQueueKey())
	if length.Val() != 1 {
		t.Errorf("Expected %d but got %d", 1, length.Val())
	}

	err = q.Ack("animal")
	if err != nil {
		t.Errorf("Expected %s but got %s", ErrNotFound, err)
	}

	err = q.Ack("drteeth")
	if err != nil {
		t.Errorf("Expected %s but got %s", ErrNotFound, err)
	}

	length = redisConn.LLen(q.procQueueKey())
	if length.Val() != 0 {
		t.Errorf("Expected %d but got %d", 0, length.Val())
	}
}

func TestGracefulShutdown(t *testing.T) {
	q := NewQueue()
	defer tearDown(q)

	q.Queue("drteeth")
	q.Queue("floyd")

	q.Dequeue()
	q.Dequeue()

	q.gracefulShutdown()

	length := redisConn.LLen(q.procQueueKey())
	if length.Val() != 0 {
		t.Errorf("Expected %d but got %d", 0, length.Val())
	}

	length = redisConn.LLen(q.waitingQueueKey())
	if length.Val() != 2 {
		t.Errorf("Expected %d but got %d", 2, length.Val())
	}

}
