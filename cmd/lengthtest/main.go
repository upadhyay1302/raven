package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/upadhyay1302/raven"
)

func sleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func main() {
	log := raven.New(raven.Auto)
	defer log.Close()

	const workerCount = 3
	log.Info("spawning threads", raven.Int("count", workerCount))

	// a subtle palette used for anchored worker lines
	// so they don't compete visually with main thread output
	subtlePalette := raven.Palette{
		raven.Transient: {Primary: raven.DarkCyan, Secondary: raven.DarkGray},
		raven.Verbose:   {Primary: raven.DarkCyan, Secondary: raven.DarkGray},
		raven.Info:      {Primary: raven.DarkCyan, Secondary: raven.DarkGray},
		raven.Warning:   {Primary: raven.DarkCyan, Secondary: raven.DarkGray},
		raven.Error:     {Primary: raven.DarkCyan, Secondary: raven.DarkGray},
	}

	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		id := i
		anchor := raven.AddAnchor(log)
		styled := raven.WithOptions(anchor, raven.OptPalette(subtlePalette))

		go func() {
			defer wg.Done()
			defer raven.RemoveAnchor(anchor)
			runWorker(styled, id)
			styled.Info("thread finished", raven.Int("thread", id))
		}()
	}

	// main thread logs long lines to test cropping behaviour
	sleep(800)
	log.Info("main thread reporting in with a really, really long line that just goes on and on and on and on and never seems to end because it keeps going and going and going and oh wow it is still going right up until it stops")
	sleep(400)
	log.Warning("main thread again, this time with a warning line that is also unusually long and full of clauses that just keep appending themselves without really ever stopping for maybe a punctuation mark or whatever but no this line just keeps going and going")
	sleep(500)
	log.Info("more from main thread, but this line is a lot shorter")

	wg.Wait()
	log.Info("done!")
}

// runWorker logs intentionally long transient lines to test how Raven
// crops them at the terminal width without wrapping.
func runWorker(log raven.Logger, id int) {
	log.Info("thread started", raven.Int("thread", id))

	for pct := 0; pct <= 100; pct++ {
		log.Transient(
			"🎃🎃 status that is really, really long; so long that it will probably fall off the end of the terminal edge... which is exactly what we want to test...",
			raven.Int("thread", id),
			raven.Int("percent", pct),
		)
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
	}
}