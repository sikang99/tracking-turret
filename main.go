package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"flag"

	"github.com/sikang99/tracking-turret/detector"
	"github.com/sikang99/tracking-turret/turret"
	"github.com/sikang99/tracking-turret/window"
)

const minArea = 7000

var (
	area   = flag.Float64("area", minArea, "base area for motion detection")
	device = flag.Int("device", 0, "device ID for the camera")
)

func main() {
	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	t, err := turret.New("33", "35", 1.3, 500, 0)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	wm := window.New(800, 600)

	detector, err := detector.New(*device, *area, t.HandleMotion, wm)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go detector.Run(ctx)
	<-c
	cancel()
	wm.Close()
}
