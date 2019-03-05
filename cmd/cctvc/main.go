package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mattn/go-mjpeg"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/darkwrat/cctvd/cctv"
)

var (
	addr  = flag.String("addr", "127.0.0.1:50051", "cctvd host:port")
	delay = flag.Duration("delay", 1*time.Second, "cctvd reconnect delay after failure")
)

func live(m map[int32]*mjpeg.Stream) error {
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("could not connect: %s", err)
	}
	defer conn.Close()

	c := cctv.NewCCTVClient(conn)
	feeds, err := c.Feeds(context.Background(), &cctv.Channels{Mask: 0xffff})
	if err != nil {
		return fmt.Errorf("could not subscribe: %s", err)
	}

	for {
		frame, err := feeds.Recv()
		if err != nil {
			return fmt.Errorf("could not receive frame: %s", err)
		}

		if frame.Channel < 32 {
			if err := m[frame.Channel].Update(frame.Image); err != nil {
				return fmt.Errorf("could not update frame for channel `%d': %s", frame.Channel, err)
			}
		}
	}
}

func main() {
	flag.Parse()

	m := make(map[int32]*mjpeg.Stream, 32)
	for i := int32(0); i < 32; i++ {
		m[i] = mjpeg.NewStream()
		pattern := fmt.Sprintf("/channel/%d/", i)
		http.Handle(pattern, m[i])
	}

	go func() {
		log.Fatal(http.ListenAndServe(":8000", nil))
	}()

	for {
		if err := live(m); err != nil {
			log.Print(err)
		}

		log.Printf("sleeping for %v seconds before retry", delay.Seconds())
		time.Sleep(*delay)
	}
}
