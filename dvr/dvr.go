package dvr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

const (
	packetHeaderMagic = uint16(0x2818)
)

type command uint16

const (
	commandUserLogonReq       = command(4115)
	commandUserLogonRes       = command(4116)
	commandSysAliveRes        = command(4160)
	commandSysInfoGetReqRes   = command(12033)
	commandCameraNameGetReq   = command(12049)
	commandCameraNameGetRes   = command(12050)
	commandMultiLiveStartReq  = command(16640)
	commandMultiLiveStartRes  = command(16641)
	commandMultiLiveStopReq   = command(16642)
	commandMultiLiveStopRes   = command(16643)
	commandMultiStreamDataRes = command(16644)
	commandMultiChStatusRes   = command(16645)
	commandMultiChgChMaskReq  = command(16646)
	commandMultiChgChMaskRes  = command(16647)
)

type packetHeader struct {
	// magic
	cmd command
	len uint16
}

func readPacketHeader(r io.Reader) (*packetHeader, error) {
	buf := make([]byte, 6)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("could not read packet header: %s", err)
	}

	magic := binary.BigEndian.Uint16(buf[:2])
	if magic != packetHeaderMagic {
		return nil, fmt.Errorf("magic mismatch: `%d' != `%d'", magic, packetHeaderMagic)
	}

	header := &packetHeader{
		cmd: command(binary.BigEndian.Uint16(buf[2:4])),
		len: binary.BigEndian.Uint16(buf[4:6]),
	}

	return header, nil
}

func readPacketBody(r io.Reader, body []byte) error {
	if _, err := io.ReadFull(r, body); err != nil {
		return fmt.Errorf("could not read packet body: %s", err)
	}

	return nil
}

type packet struct {
	header *packetHeader
	body   []byte
}

func readPacket(r io.Reader) (*packet, error) {
	header, err := readPacketHeader(r)
	if err != nil {
		return nil, err
	}

	pkt := &packet{header: header}
	if header.len > 6 {
		pkt.body = make([]byte, header.len-6)
		if err := readPacketBody(r, pkt.body); err != nil {
			return nil, err
		}
	}

	return pkt, nil
}

func writePacketHeader(w io.Writer, header *packetHeader) error {
	buf := make([]byte, 6)
	binary.BigEndian.PutUint16(buf[0:2], packetHeaderMagic)
	binary.BigEndian.PutUint16(buf[2:4], uint16(header.cmd))
	binary.BigEndian.PutUint16(buf[4:6], uint16(header.len))

	if _, err := w.Write(buf); err != nil {
		return fmt.Errorf("could not write packet header: %s", err)
	}

	return nil
}

func writePacketBody(w io.Writer, body []byte) error {
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("could not write packet body: %s", err)
	}

	return nil
}

func writePacket(w io.Writer, cmd command, data []byte) error {
	header := &packetHeader{
		cmd: cmd,
		len: uint16(len(data) + 6),
	}

	if err := writePacketHeader(w, header); err != nil {
		return err
	}
	if err := writePacketBody(w, data); err != nil {
		return err
	}

	return nil
}

func waitPacket(r io.Reader, cmd command) (*packet, error) {
	for {
		pkt, err := readPacket(r)
		if err != nil {
			return nil, err
		}
		if pkt.header.cmd == cmd {
			return pkt, nil
		}
	}
}

func makeAuthPacketBody(user, pass string) []byte {
	var body []byte
	body = append(body, []byte(user)[:32]...)
	body = append(body, []byte(pass)[:32]...)
	return body
}

func checkAuthResult(r uint16) error {
	switch r {
	case 61441, 61442, 61443, 61444, 61445, 61446, 65535:
		return fmt.Errorf("reason `%d'", r)
	default:
		return nil
	}
}

func doAuth(conn net.Conn, user, password string) error {
	buf := makeAuthPacketBody(user, password)
	if err := writePacket(conn, commandUserLogonReq, buf); err != nil {
		return err
	}

	response, err := waitPacket(conn, commandUserLogonRes)
	if err != nil {
		return err
	}

	result := binary.BigEndian.Uint16(response.body)
	if err := checkAuthResult(result); err != nil {
		return err
	}

	return nil
}

func checkSysInfoGetResult(r uint16) error {
	switch r {
	case 0:
		return nil
	default:
		return fmt.Errorf("reason `%d'", r)
	}
}

type sysInfo struct {
	channelCount uint16
}

func doSysInfoGet(conn net.Conn) (*sysInfo, error) {
	if err := writePacket(conn, commandSysInfoGetReqRes, nil); err != nil {
		return nil, err
	}

	response, err := waitPacket(conn, commandSysInfoGetReqRes)
	if err != nil {
		return nil, err
	}

	if err := checkSysInfoGetResult(binary.BigEndian.Uint16(response.body)); err != nil {
		return nil, err
	}

	return &sysInfo{
		channelCount: binary.BigEndian.Uint16(response.body[34:]),
	}, nil
}

func checkChannelsGetResult(r uint16) error {
	switch r {
	case 0:
		return nil
	default:
		return fmt.Errorf("reason `%d'", r)
	}
}

func doChannelsGet(conn net.Conn, n uint16) (map[uint8]string, error) {
	if err := writePacket(conn, commandCameraNameGetReq, nil); err != nil {
		return nil, err
	}

	response, err := waitPacket(conn, commandCameraNameGetRes)
	if err != nil {
		return nil, err
	}

	p := response.body
	if err := checkChannelsGetResult(binary.BigEndian.Uint16(p)); err != nil {
		return nil, err
	}

	p = p[4:]
	chs := make(map[uint8]string, n)
	for i := uint8(0); i < uint8(n); i++ {
		chs[i] = string(p[:32])
		p = p[32:]
	}

	return chs, nil
}

type Client struct {
	conn     net.Conn
	info     *sysInfo
	channels map[uint8]string
}

type ConnectOpts struct {
	Addr, User, Password string
}

func Connect(opts ConnectOpts) (*Client, error) {
	conn, err := net.Dial("tcp", opts.Addr)
	if err != nil {
		return nil, fmt.Errorf("could not dial: %s", err)
	}

	if err := doAuth(conn, opts.User, opts.Password); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("auth failed: %s", err)
	}

	info, err := doSysInfoGet(conn)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("sys info get failed: %s", err)
	}

	channels, err := doChannelsGet(conn, info.channelCount)
	if err != nil {
		return nil, fmt.Errorf("could not get channels: %s", err)
	}

	return &Client{conn: conn, info: info, channels: channels}, nil
}

func (c *Client) Close() {
	_ = c.conn.Close()
}

func (c *Client) Channels() map[uint8]string {
	return c.channels
}

func (c *Client) doStartLive() error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf[0:2], 0xffff)
	binary.BigEndian.PutUint16(buf[2:4], 1)
	return writePacket(c.conn, commandMultiLiveStartReq, buf)
}

func (c *Client) doChangeMask(mask uint16) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, mask)
	return writePacket(c.conn, commandMultiChgChMaskReq, buf) // skip response
}

type streamPacketInfo struct {
	channel     uint16
	frameTime   uint32
	fragTotal   uint8
	fragCurrent uint8
}

func unpackStreamPacketInfo(b []byte) streamPacketInfo {
	f := binary.BigEndian.Uint16(b[20:22])
	return streamPacketInfo{
		channel:     binary.BigEndian.Uint16(b[2:4]),
		frameTime:   binary.BigEndian.Uint32(b[16:20]),
		fragTotal:   uint8(f >> 8),
		fragCurrent: uint8(f & 0xff),
	}
}

type Frame struct {
	Channel uint16
	Ts      uint32
	Image   []byte
}

type frameBuilder struct {
	ch  chan *Frame
	buf *bytes.Buffer
}

func (b *frameBuilder) update(info streamPacketInfo, pkt *packet) error {
	p := pkt.body[24:]
	if info.fragCurrent == 0 {
		p = p[32:]
		b.buf.Reset()
	}

	if _, err := io.CopyN(b.buf, bytes.NewReader(p), int64(len(p))); err != nil {
		return fmt.Errorf("could not copy: %s", err)
	}

	if info.fragCurrent+1 == info.fragTotal {
		b.ch <- &Frame{
			Channel: info.channel,
			Ts:      info.frameTime,
			Image:   b.buf.Bytes(),
		}
	}

	return nil
}

func (c *Client) Live(ch chan *Frame) error {
	defer c.Close()

	m := make(map[uint16]*frameBuilder)
	for i := uint16(0); i < c.info.channelCount; i++ {
		m[i] = &frameBuilder{ch: ch, buf: &bytes.Buffer{}}
	}

	if err := c.doStartLive(); err != nil {
		return fmt.Errorf("could not start live: %s", err)
	}

	if err := c.doChangeMask(0xffff); err != nil {
		return fmt.Errorf("could not change mask: %s", err)
	}

	ping := time.NewTicker(10 * time.Second)
	defer ping.Stop()

	for {
		pkt, err := readPacket(c.conn)
		if err != nil {
			log.Printf("could not read packet: %s", err)
			continue
		}

		select {
		case <-ping.C:
			_ = writePacket(c.conn, commandSysAliveRes, nil)
		default:
		}

		switch pkt.header.cmd {
		case commandMultiStreamDataRes:
			info := unpackStreamPacketInfo(pkt.body)
			if b, ok := m[info.channel]; ok {
				if err := b.update(info, pkt); err != nil {
					log.Printf("could not build frame: %s", err)
					continue
				}
			}
		}
	}
}
