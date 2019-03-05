package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
	"time"

	"github.com/darkwrat/cctvd/cctv"
	"github.com/darkwrat/cctvd/dvr"
)

type cs map[chan *cctv.Frame]struct{}

type server struct {
	sync.RWMutex
	m map[uint8]cs
}

func (s *server) subscribe(mask int32, sub chan *cctv.Frame) {
	s.Lock()
	defer s.Unlock()

	for feed := uint8(0); feed < 32; feed++ {
		if v := mask & (1 << feed); v == 0 {
			continue
		}
		if c, ok := s.m[feed]; ok {
			c[sub] = struct{}{}
		} else {
			s.m[feed] = cs{sub: {}}
		}
	}
}

func (s *server) unsubscribe(mask int32, sub chan *cctv.Frame) {
	s.Lock()
	defer s.Unlock()

	for ch := uint8(0); ch < 32; ch++ {
		if v := mask & (1 << ch); v != 0 {
			if c, ok := s.m[ch]; ok {
				delete(c, sub)
			}
		}
	}
}

func (s *server) send(frame *cctv.Frame) {
	s.RLock()
	defer s.RUnlock()

	// todo fine grained lock
	// todo handle client drop
	ch := uint8(frame.GetChannel())
	if c, ok := s.m[ch]; ok {
		for sub := range c {
			sub <- frame
		}
	}
}

func (s *server) Feeds(r *cctv.Channels, stream cctv.CCTV_FeedsServer) error {
	mask := r.GetMask()
	if mask == 0 {
		return nil
	}

	sub := make(chan *cctv.Frame, 100)
	s.subscribe(mask, sub)
	defer func() {
		s.unsubscribe(mask, sub)
		close(sub)
	}()

	for frame := range sub {
		if err := stream.Send(frame); err != nil {
			log.Printf("could not send frame: %s", err)
			break
		}
	}

	return nil
}

func (s *server) multicast(ch chan *dvr.Frame) {
	for f := range ch {
		s.send(&cctv.Frame{
			Channel: int32(f.Channel),
			Ts:      int64(f.Ts),
			Image:   f.Image,
		})
	}
}

var (
	addr  = flag.String("addr", "127.0.0.1:7620", "dvr host:port")
	delay = flag.Duration("delay", 5*time.Second, "delay before relive after failure")
)

func live(opts dvr.ConnectOpts, ch chan *dvr.Frame) error {
	c, err := dvr.Connect(opts)
	if err != nil {
		return fmt.Errorf("cannot connect to dvr: %s", err)
	}
	defer c.Close()

	if err := c.Live(ch); err != nil {
		return fmt.Errorf("cannot stream anymore: %s", err)
	}

	return nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	s := grpc.NewServer()
	csrv := &server{m: make(map[uint8]cs)}
	cctv.RegisterCCTVServer(s, csrv)

	ch := make(chan *dvr.Frame, 100)
	go csrv.multicast(ch)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("cannot serve grpc: %s", err)
		}
	}()

	opts := dvr.ConnectOpts{
		Addr:     *addr,
		User:     "ADMIN",
		Password: "0000",
	}

	for {
		if err := live(opts, ch); err != nil {
			log.Print(err)
		}

		log.Printf("sleeping for %v seconds before retry", delay.Seconds())
		time.Sleep(*delay)
	}
}
