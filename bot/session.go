package bot

import (
	"context"
	"time"

	"github.com/Richtermnd/anecdotebot/client"
)

type session struct {
	ID       int64
	Category int           `json:"category"`
	Delay    time.Duration `json:"delay"`
	cancel   func()        `json:"-"`
}

func Session(id int64) *session {
	s, ok := sessions[id]
	if ok {
		return s
	}
	s = &session{
		ID:       id,
		Category: 1,
		Delay:    time.Hour,
	}
	sessions[id] = s
	return s
}

func (s *session) SendAnecdote() {
	log := log.With("id", s.ID)
	log.Debug("send anecdote")
	anecdote := client.GetAnecdote(s.Category)
	if anecdote == "" {
		SendText(s.ID, "Что-то пошло не так, но в этом виноват не я. Попробуйте ещё раз.")
	}
	SendText(s.ID, anecdote)
	log.Debug("anecdote sent")
}

func (s *session) Notify() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go func(ctx context.Context) {
		for {
			log := log.With("id", s.ID)
			select {
			case <-ctx.Done():
				return
			default:
				log.Debug("send anecdote")
				s.SendAnecdote()
			}
			time.Sleep(s.Delay)
		}
	}(ctx)
}

func (s *session) StopNotify() {
	if s.cancel == nil {
		return
	}
	s.cancel()
}
