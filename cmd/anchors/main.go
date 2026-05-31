package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/upadhyay1302/raven"
)

func main() {
	log := raven.New(raven.Auto)
	defer log.Close()

	log.Info("spawning worker threads", raven.Int("count", 3))
	time.Sleep(time.Second)

	var wg sync.WaitGroup
	const workerCount = 3
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		id := i
		go func() {
			defer wg.Done()
			doWork(log, id)
			doWork(log, id)
		}()
	}

	time.Sleep(time.Second)
	log.Info("waited one second...")
	time.Sleep(time.Second)
	log.Warning("waited two seconds...")
	time.Sleep(time.Second)
	log.Error("bored of waiting")

	wg.Wait()
	log.Info("all threads done!")
}

// doWork simulates a unit of work with a live anchored progress line.
// The anchor is created and removed within this function so the pinned
// line disappears cleanly when the work is complete.
func doWork(parent raven.Logger, id int) {
	log := raven.AddAnchor(parent)
	defer raven.RemoveAnchor(log)

	log.Transient("starting...", raven.Int("thread", id))
	time.Sleep(time.Duration(400*id) * time.Millisecond)

	for pct := 0; pct <= 100; pct++ {
		log.Transient("working",
			raven.Int("thread", id),
			raven.Int("percent", pct),
		)
		time.Sleep(time.Duration(5+rand.Intn(30)) * time.Millisecond)
	}
}