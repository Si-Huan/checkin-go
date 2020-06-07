package cks

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

var ctx = context.Background()

type Cks struct {
	logger  *logrus.Logger
	Server  *Server
	RedisDB *RedisDB
	Timer   *Timer
}

func NewCks() *Cks {
	c := new(Cks)

	c.logger = logrus.New()

	c.Timer = NewTimer(c)
	c.Server = NewServer(c)
	c.RedisDB = NewRedisDB(c)

	return c
}

func (c *Cks) Start() {

	switch os.Getenv("CKSFIRSTRUN") {
	case "true":
		c.RedisDB.LoadData()
		break
	}

	c.Server.Start()
	c.Timer.Start()
}
