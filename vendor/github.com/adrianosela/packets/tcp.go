package packets

import (
	"encoding/binary"
)

const (
	tcpMinHeaderLen = 20
)

// TCPHeader represents the fields in a tcp header
type TCPHeader struct {
	Src     int
	Dst     int
	Seq     int
	Ack     int
	Len     int
	Rsvd    int
	Flag    int
	Win     int
	Sum     int
	Urp     int
	Options []byte
}

// RawTCPHeader encodes a tcp header
// as per https://tools.ietf.org/html/rfc793
func RawTCPHeader(p *TCPHeader) []byte {
	headerLen := tcpMinHeaderLen + len(p.Options)
	raw := make([]byte, headerLen)

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |          Source Port          |       Destination Port        |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	binary.BigEndian.PutUint16(raw[0:2], uint16(p.Src))
	binary.BigEndian.PutUint16(raw[2:4], uint16(p.Dst))

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                        Sequence Number                        |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                    Acknowledgment Number                      |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	binary.BigEndian.PutUint32(raw[4:8], uint32(p.Seq))
	binary.BigEndian.PutUint32(raw[8:12], uint32(p.Ack))

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |  Data |           |U|A|P|R|S|F|                               |
	// | Offset| Reserved  |R|C|S|S|Y|I|            Window             |
	// |       |           |G|K|H|T|N|N|                               |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	raw[12] = uint8(headerLen/4<<4 | 0) //TODO:  Reserved
	raw[13] = uint8(p.Flag)
	binary.BigEndian.PutUint16(raw[14:16], uint16(p.Win))

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |           Checksum            |         Urgent Pointer        |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	binary.BigEndian.PutUint16(raw[16:18], uint16(p.Sum))
	binary.BigEndian.PutUint16(raw[18:20], uint16(p.Urp))

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                    Options                    |    Padding    |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	if len(p.Options) > 0 {
		copy(raw[tcpMinHeaderLen:], p.Options)
	}

	// data begins after padding

	return raw
}
