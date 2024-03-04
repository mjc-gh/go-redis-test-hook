package redistesthooks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setup() (*redis.Client, *Hook) {
	hook := New()

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	rdb.AddHook(hook)

	return rdb, hook
}

type expectations struct {
	name string
	fn   func(context.Context, *redis.Client)
	cmds []string
}

func TestHookProcessHook(t *testing.T) {
	ctx := context.TODO()
	rdb, hook := setup()

	t1 := time.Now().UnixNano()
	t2 := time.Now().UnixNano()

	var cases = []expectations{
		expectations{
			name: "basic",
			fn: func(ctx context.Context, rdb *redis.Client) {
				rdb.Set(ctx, "key", 42, time.Duration(0))
				rdb.Get(ctx, "key")
			},
			cmds: []string{"SET key 42", "GET key"},
		},
		expectations{
			name: "complex args",
			fn: func(ctx context.Context, rdb *redis.Client) {
				rdb.HSet(ctx, "hash", map[string]interface{}{"key1": "value1", "key2": "value2"})
				rdb.ZAdd(ctx, "oset",
					redis.Z{float64(t1), "a"},
					redis.Z{float64(t2), "b"},
				)
			},
			cmds: []string{
				"HSET hash key1 value1 key2 value2",
				fmt.Sprintf("ZADD oset %d a %d b", t1, t2),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.fn(ctx, rdb)

			assert.Len(t, hook.Captures, len(tc.cmds))
			for i := range tc.cmds {
				assert.Equal(t, tc.cmds[i], hook.Captures[i].String())
			}

			hook.Reset()
			assert.Empty(t, hook.Captures)
		})
	}
}
