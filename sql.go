package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	// TODO: Missing a lot of the capability flags, only included the ones relevant to decoding
	clientPluginAuth       = 0x00080000
	clientSecureConnection = 0x00008000
)

// MySQLv10 is the MySQL v10 handshake packet
// This packet is described here:
// https://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::Handshake
type MySQLv10 struct {
	// ServerVersion in human readable version
	ServerVersion string

	// ConnectionId from the handshake packet, not sure if this is useful
	ConnectionId uint32

	// CharacterSet default character set, this is  collation ID in the table from the link
	// https://dev.mysql.com/doc/internals/en/character-set.html#packet-Protocol::CharacterSet
	CharacterSet uint8

	// Status is a bit-field of status flags described here:
	// https://dev.mysql.com/doc/internals/en/status-flags.html#packet-Protocol::StatusFlags
	// Referred to as status_flags in the handshake doc
	Status uint16

	// Capabilities are the capability flags described here:
	// https://dev.mysql.com/doc/internals/en/capability-flags.html#packet-Protocol::CapabilityFlags
	// Combined capability_flags_1 and capability_flags_2 (if capability_flags_2 existed) from handshake doc
	Capabilities uint32

	// AuthPlugin is the name of the authentication method
	// Referred to as auth_plugin_name in the handshake doc
	AuthPlugin string

	// AuthData is the combined auth plugin data
	// Referred to as auth_plugin_data_part_1 and auth_plugin_data_part_2 from handshake doc
	// This is commonly called the Cipher or Salt, but depends on the auth plugin
	AuthData []byte
}

var (
	ErrorMissingData     = errors.New("Not enough data received for MySQLv10 handshake")
	ErrorInvalidProtocol = errors.New("MySQL Handshake version doesn't match expected")
)

// DetectMySQL on the given host
// Use timeout parameter when dialing connection
func DetectMySQL(host string, timeout int) (*MySQLv10, error) {
	conn, err := net.DialTimeout("tcp", host, time.Second*time.Duration(timeout))
	if err != nil {
		return nil, fmt.Errorf("Failed to detect MySQL during connect: %s\n", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	if _, err := conn.Read(buf); err != nil {
		return nil, fmt.Errorf("Failed to detect MySQL during read: %s\n", err)
	}

	sql := MySQLv10{}
	if err = sql.Decode(buf); err != nil {
		return nil, fmt.Errorf("Failed to detect MySQL during decode: %s\n", err)
	}

	return &sql, nil
}

// String output to a human readable form
// TODO: Add all the capabilities to this and print values as hex
func (s *MySQLv10) String() string {
	return fmt.Sprintf("%+v", *s)
}

// Decode the handshake packet given the byte slice
// Handshake packet described here:
// https://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::Handshake
//
// Another usage reference in connector.go:
// https://github.com/go-sql-driver/mysql
func (s *MySQLv10) Decode(buf []byte) error {
	if len(buf) < 4 {
		return ErrorMissingData
	}

	// First 3 bytes are the packet length of the handshake packet
	pktLen := int(uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16)

	// There is another byte representing the sequence, but it doesn't seem useful
	// seq := buf[3]

	if pktLen+4 > len(buf) {
		return ErrorMissingData
	}

	// Start using position variable to keep track of decoding
	pos := 4

	// protocol_version(1) This is only meant to work with version 10
	if 10 != buf[pos] {
		return ErrorInvalidProtocol
	}
	pos += 1

	// server_version(null terminated string)
	s.ServerVersion = read_cstr(buf[pos:])
	pos += len(s.ServerVersion) + 1 // Extra +1 for the null terminator

	// connection_id(4)
	s.ConnectionId = binary.LittleEndian.Uint32(buf[pos : pos+4])
	pos += 4

	// auth_plugin_data_1(8) 8 byte string representing the first 8 bytes of auth-plugin data
	authData := buf[pos : pos+8]
	pos += 8 + 1 // Extra +1 because of filler_1(1) which is just a zeroed byte

	// capability_flag_1(2) lower two bytes of the capabilities flags
	s.Capabilities = uint32(binary.LittleEndian.Uint16(buf[pos : pos+2]))
	pos += 2

	// If there are still more data within the packet we have more "extended fields"
	if pos < pktLen+4 {
		// character_set(1)
		s.CharacterSet = buf[pos]
		pos += 1

		// status_flags(2) bit-fields representing status
		s.Status = binary.LittleEndian.Uint16(buf[pos : pos+2])
		pos += 2

		// capability_flags_2(2) upper two bytes of the capabilities flags sometimes called extended capabilities
		s.Capabilities |= uint32(binary.LittleEndian.Uint16(buf[pos:pos+2])) << 16
		pos += 2

		// auth_data_plugin_len(1) Length of the second plugin data piece
		authLen := -1
		if s.Capabilities&clientPluginAuth != 0 {
			authLen = int(buf[pos])
		}
		pos += 1 + 10 // Extra +10 for a reserved section, this should be zeroed out

		if s.Capabilities&clientSecureConnection != 0 {
			// Remaining auth data length is described on dev.mysql.com as max(13, auth_data_plugin_len - 8)
			authDataLen := 13
			if authLen-8 > authDataLen {
				authDataLen = authLen - 8
			}
			authDataLen -= 1 // Last byte is null so just remove it

			// auth_plugin_data_part_2(authDataLen) second part of the cipher
			authData = append(authData, buf[pos:pos+authDataLen]...)
			pos += authDataLen + 1 // Add the null byte back
		}

		if s.Capabilities&clientPluginAuth != 0 {
			// auth_plugin_name(null terminated string) name of the auth method
			s.AuthPlugin = read_cstr(buf[pos:])
		}
	}

	s.AuthData = make([]byte, len(authData))
	copy(s.AuthData, authData)
	return nil
}

// Read a null terminated string from a byte slice
func read_cstr(buf []byte) string {
	pos := bytes.IndexByte(buf, 0)
	if pos == -1 {
		return string(buf[:])
	}

	return string(buf[:pos])
}
