package packets

import (
	"encoding/binary"
	"net"
)

const (
	ipVersion        = 4
	ipv4MinHeaderLen = 20
)

// IPv4Header represents the fields in an IPv4 header
type IPv4Header struct {
	Version  int
	Len      int
	TOS      int
	TotalLen int
	ID       int
	Flags    int
	FragOff  int
	TTL      int
	Protocol int
	Checksum int
	Src      net.IP
	Dst      net.IP
	Options  []byte
}

// RawIPv4Header encodes an IPv4 header
// as per https://tools.ietf.org/html/rfc791
func RawIPv4Header(p *IPv4Header) []byte {
	headerLen := ipv4MinHeaderLen + len(p.Options)
	raw := make([]byte, headerLen)

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |Version|  IHL  |Type of Service|          Total Length         |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	raw[0] = byte(ipVersion<<4 | (headerLen >> 2 & 0x0f))
	raw[1] = byte(p.TOS)
	binary.BigEndian.PutUint16(raw[2:4], uint16(p.TotalLen))

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |         Identification        |Flags|      Fragment Offset    |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	binary.BigEndian.PutUint16(raw[4:6], uint16(p.ID))
	flagsAndFragOff := (p.FragOff & 0x1fff) | int(p.Flags<<13)
	binary.BigEndian.PutUint16(raw[6:8], uint16(flagsAndFragOff))

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |  Time to Live |    Protocol   |         Header Checksum       |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	raw[8] = byte(p.TTL)
	raw[9] = byte(p.Protocol)
	binary.BigEndian.PutUint16(raw[10:12], uint16(p.Checksum))

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                       Source Address                          |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                    Destination Address                        |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	if ip := p.Src.To4(); ip != nil {
		copy(raw[12:16], ip[:net.IPv4len])
	}
	if ip := p.Dst.To4(); ip != nil {
		copy(raw[16:20], ip[:net.IPv4len])
	}

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                    Options                    |    Padding    |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	if len(p.Options) > 0 {
		copy(raw[ipv4MinHeaderLen:], p.Options)
	}

	// data begins after padding

	return raw
}
