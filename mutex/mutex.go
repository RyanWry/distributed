package mutex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

var delScript = redis.NewScript(1, `
if (redis.call('get', KEYS[1]) == ARGV[1])
		then return redis.call('del', KEYS[1])
else return 0
`)

type Mutex struct {
	id     string
	key    string
	ttl    int64
	cancel context.CancelFunc
	conn   redis.Conn
}

func NewMutex(key string, conn redis.Conn) *Mutex {
	return &Mutex{
		id:  uuid.New().String(),
		key: key,
		ttl: 30,
	}
}

func (m *Mutex) TryLock() error {
	res, err := redis.String(m.conn.Do("SET", m.key, m.id, "NX", "EX", m.ttl))
	if err != nil {
		return err
	}
	if res != "OK" {
		return fmt.Errorf("lock error: %s", res)
	}

	ctx, cancle := context.WithCancel(context.TODO())
	m.cancel = cancle

	go m.renew(ctx)
	return nil
}

func (m *Mutex) UnLock() error {
	if m.cancel != nil {
		m.cancel()
	}

	res, err := redis.Int(delScript.Do(m.conn, m.key, m.id))
	if err != nil {
		return err
	}
	if res != 1 {
		return fmt.Errorf("del error: %d", res)
	}
	return nil
}

func (m *Mutex) renew(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := m.conn.Do("EXPIRE", m.key, m.ttl)
			if err != nil {
				log.Println("renew error: ", err)
			}
			time.Sleep(time.Duration(m.ttl/3) * time.Second)
		}
	}
}
