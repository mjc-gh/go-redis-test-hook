package redistesthooks

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/redis/go-redis/extra/rediscmd/v9"
	"github.com/redis/go-redis/v9"
)

type CmdCap struct {
	Name string
	Args []string
}

func (cc CmdCap) String() string {
	return fmt.Sprintf("%s %s", strings.ToUpper(cc.Name), strings.Join(cc.Args, " "))
}

type Hook struct {
	Captures []CmdCap
}

func New() *Hook {
	return &Hook{Captures: make([]CmdCap, 0)}
}

func (h *Hook) Reset() {
	h.Captures = make([]CmdCap, 0)
}

func (h *Hook) DialHook(hook redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return hook(ctx, network, addr)
	}
}

func (h *Hook) ProcessHook(hook redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		h.Captures = append(h.Captures, newCmdCap(cmd))

		return hook(ctx, cmd)
	}
}

func (h *Hook) ProcessPipelineHook(hook redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		for _, cmd := range cmds {
			h.Captures = append(h.Captures, newCmdCap(cmd))
		}

		return hook(ctx, cmds)
	}
}

func newCmdCap(cmd redis.Cmder) (cc CmdCap) {
	cmdStr := rediscmd.CmdString(cmd)
	for idx, str := range strings.Split(cmdStr, " ") {
		if idx == 0 {
			cc.Name = str
		} else {
			cc.Args = append(cc.Args, str)
		}
	}

	return
}
