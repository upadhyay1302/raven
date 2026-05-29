package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/upadhyay1302/raven"
)

var (
	verbose  = flag.Bool("verbose", false, "drop threshold from info to verbose")
	useJSON  = flag.Bool("json", false, "output structured JSON")
	noTime   = flag.Bool("notime", false, "do not include timestamps (ignored with -json)")
	noLevel  = flag.Bool("nolevel", false, "do not include level label (ignored with -json)")
	swapLayout = flag.Bool("swap", false, "swap message and fields in output (ignored with -json)")
)

func main() {
	flag.Parse()

	style := raven.Auto
	if *useJSON {
		style = raven.JSON
	}

	var layout raven.PrinterOption = raven.OptMsgThenFields
	if *swapLayout {
		layout = raven.OptFieldsThenMsg
	}

	log := raven.New(style,
		raven.OptColumnOffset(40),
		raven.OptShowTime(!*noTime),
		raven.OptShowLevel(!*noLevel),
		layout,
	)
	defer log.Close()

	log.Info("Raven Demo App")

	// print all available flags
	flag.VisitAll(func(f *flag.Flag) {
		log.Info(fmt.Sprintf("  --%s :: %s", f.Name, f.Usage))
	})

	// log all command line args as structured fields
	var argFields []raven.Fielder
	for i, v := range os.Args {
		arg := v
		if i == 0 {
			arg = filepath.Base(arg)
		}
		argFields = append(argFields, raven.String(fmt.Sprintf("arg%d", i), arg))
	}
	log.Info("os.Args", argFields...)

	// show all log levels
	log.SetThreshold(raven.Transient)
	log.Transient("transient line")
	log.Verbose("verbose line")
	log.Info("info line")
	log.Warning("warning line")
	log.Error("error line")

	// set working threshold
	if *verbose {
		log.SetThreshold(raven.Verbose)
	} else {
		log.SetThreshold(raven.Info)
	}

	// spawn worker threads with anchored progress lines
	threads := 5
	log.Info("spawning worker threads", raven.Int("count", threads))

	var wg sync.WaitGroup
	wg.Add(threads)

	for i := 0; i < threads; i++ {
		n := i
		anchor := raven.AddAnchor(log)
		go func() {
			defer wg.Done()
			defer raven.RemoveAnchor(anchor)
			anchor.Verbose("thread spawned", raven.Int("thread", n))
			runWorker(anchor, n)
			anchor.Verbose("thread closing", raven.Int("thread", n))
		}()
	}

	// main thread activity while workers run
	time.Sleep(time.Second)
	log.Info("still running...")
	time.Sleep(500 * time.Millisecond)
	log.Info("yup, still running...")
	time.Sleep(100 * time.Millisecond)
	log.Warning("something happened on the main thread")
	time.Sleep(500 * time.Millisecond)
	log.Info("main thread checking in")
	time.Sleep(5 * time.Second)
	log.Error("main thread encountered an error")

	wg.Wait()
	log.Info("all done!")
}

// runWorker simulates a download process with progress updates.
func runWorker(log raven.Logger, id int) {
	log.Transient("starting...", raven.Int("thread", id))
	time.Sleep(time.Duration(400*id) * time.Millisecond)

	for pct := 0; pct <= 100; pct++ {
		switch pct {
		case 90:
			log.Verbose("transitioning to write phase", raven.Int("thread", id))
		case 100:
			log.Info("download complete", raven.Int("thread", id))
		}

		log.Transient("downloading",
			raven.Int("thread", id),
			raven.Int("percent", pct),
		)

		sleepMs := 50 - (10 * id) + rand.Intn(50)
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)

		// randomly simulate a retry
		if pct == 50 && rand.Intn(3) == 0 {
			log.Warning("encountered a problem, retrying",
				raven.Int("thread", id),
				raven.Int("percent", pct),
			)
			time.Sleep(time.Duration(id+1) * time.Second)
		}
	}
}