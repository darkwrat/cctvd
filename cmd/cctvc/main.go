package main

import (
	"flag"
	"fmt"
	"github.com/mattn/go-mjpeg"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"time"

	"github.com/darkwrat/cctvd/cctv"
)

var (
	addr = flag.String("addr", "127.0.0.1:50051", "cctvd host:port")
)

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

outer:
	for {
		conn, err := grpc.Dial(*addr, grpc.WithInsecure())
		if err != nil {
			log.Printf("could not connect: %s", err)
			time.Sleep(1 * time.Second)
			continue outer
		}

		c := cctv.NewCCTVClient(conn)
		feeds, err := c.Feeds(context.Background(), &cctv.Channels{Mask: 0xffff})
		if err != nil {
			log.Printf("could not subscribe: %s", err)
			_ = conn.Close()
			time.Sleep(1 * time.Second)
			continue outer
		}

		for {
			frame, err := feeds.Recv()
			if err != nil {
				log.Printf("could not receive frame: %s", err)
				_ = conn.Close()
				time.Sleep(1 * time.Second)
				continue outer
			}

			if frame.Channel < 32 {
				if err := m[frame.Channel].Update(frame.Image); err != nil {
					log.Printf("could not update frame for channel `%d': %s", frame.Channel, err)
					_ = conn.Close()
					time.Sleep(1 * time.Second)
					continue outer
				}
			}
		}
	}
}
