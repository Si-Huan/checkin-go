package main

import (
	"time"

	cks "github.com/Si-Huan/checkin/lib"
)

//cks means checkin system
func main() {

	cks := cks.NewCks()

	go cks.Start()

	for {
		time.Sleep(1 * time.Minute)
	}

}
