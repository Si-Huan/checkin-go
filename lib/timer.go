package cks

import (
	"github.com/robfig/cron/v3"
)

type Timer struct {
	parent *Cks
	cron   *cron.Cron
}

func NewTimer(parent *Cks) *Timer {

	t := new(Timer)

	t.parent = parent

	t.cron = cron.New()

	t.cron.AddFunc("55 16 * * *", func() { t.parent.RedisDB.Reload() })

	return t
}

func (t *Timer) Start() {
	go t.cron.Start()
}
