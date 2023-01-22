package main

import (
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"

	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatalf("Error: %s\n", err)
	}

	defer conn.Close()

	go receive(conn)

	send(conn)
}

func receive(conn *icmp.PacketConn) {
	for {
		var b = make([]byte, 1024)
		n, r, err := conn.ReadFrom(b)
		if err != nil {
			log.Fatalf("Conn: Err: %s", err)
		}

		b = b[:n]

		msg, err := icmp.ParseMessage(1, b)
		if err != nil {
			log.Printf("recerr: %s\n", err)
			continue
		}

		if msg.Type == ipv4.ICMPTypeEchoReply {
			fmt.Print(".")
			continue
		}

		text, err := msg.Body.Marshal(1)
		if err != nil {
			log.Printf("recerr: %s\n", err)
			continue
		}

		//		log.Printf("[%s] rec -> %d bytes: %x\n", r, len(b), b)
		log.Printf("[%s] rec: %s; type: %s; code: %d\n", r, text, msg.Type, msg.Code)
	}

}

func send(conn *icmp.PacketConn) {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		log.Println(s.Text())
		split := strings.SplitN(s.Text(), " ", 2)
		if len(split) != 2 {
			log.Printf("ERROR: wrong length\n")
			continue
		}

		addr, text := split[0], split[1]

		add, err := net.ResolveIPAddr("ip4:icmp", addr)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
			continue
		}

		if len(text) < 4 {
			text += strings.Repeat(" ", 4-len(text))
		}

		msg := &icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 1337,
			Body: &body{text},
		}

		body, err := msg.Marshal(nil)
		if err != nil {
			log.Printf("Error: %s\n", err)
			continue
		}

		i, err := conn.WriteTo(body, add)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		} else {
			log.Printf("[%s] snd -> %d bytes\n", addr, i)
		}
	}
}

func Checksum(chunk []byte) uint16 {
	var sum uint32

	buf := bytes.NewBuffer(chunk)
	if len(chunk)%2 != 0 {
		buf.WriteByte(0)
	}
	for {
		var value uint16
		if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
			break
		}
		sum += uint32(value)
	}
	for {
		if s := sum >> 16; s == 0 {
			break
		}
		sum = (sum & 0xffff) | (sum >> 16)
	}
	return uint16(sum)
}

type body struct {
	str string
}

func (b *body) Len(int) int {
	return len(b.str)
}

func (b *body) Marshal(int) ([]byte, error) {
	return []byte(b.str), nil
}
