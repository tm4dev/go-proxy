package handlers

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vlourme/go-proxy/internal/auth"
	"github.com/vlourme/go-proxy/internal/config"
	"github.com/vlourme/go-proxy/internal/nio"
)

// SOCKS constants
const (
	Version4 = 0x04
	Version5 = 0x05
)

// SOCKS5 Auth Methods
const (
	AuthNoAuth       = 0x00
	AuthGSSAPI       = 0x01
	AuthUsernamePass = 0x02
	AuthNoAcceptable = 0xFF
)

// SOCKS5 Command Codes
const (
	CmdConnect = 0x01
	CmdBind
	CmdUDPAssociate
)

// SOCKS5 Address Types
const (
	AtypIPv4   = 0x01
	AtypDomain = 0x03
	AtypIPv6   = 0x04
)

// SOCKS5 Reply Codes
const (
	RepSuccess = iota
	RepGeneralFailure
	RepConnectionNotAllowed
	RepNetworkUnreachable
	RepHostUnreachable
	RepConnectionRefused
	RepTTLExpired
	RepCmdNotSupported
	RepAddrTypeNotSupported
)

// IsSocks checks if the request is a SOCKS request
func IsSocks(buf *bufio.Reader) bool {
	ver, err := buf.ReadByte()
	if err != nil {
		return false
	}
	defer buf.UnreadByte()
	return ver == Version4 || ver == Version5
}

// HandleSocks handles the SOCKS protocol
func HandleSocks(conn net.Conn, buf *bufio.Reader) int64 {
	ver, err := buf.ReadByte()
	if err != nil {
		log.Error().Err(err).Msg("failed to read version")
		return -1
	}

	if ver == Version4 {
		log.Error().Msg("socks4 is not implemented")
		return -1
	}

	return HandleSocks5(conn, buf)
}

// HandleSocks5 handles the SOCKS5 protocol
func HandleSocks5(conn net.Conn, buf *bufio.Reader) int64 {
	methodsCount, err := buf.ReadByte()
	if err != nil {
		log.Error().Err(err).Msg("failed to read methods count")
		return -1
	}

	methods := make([]byte, int(methodsCount))
	if _, err := io.ReadFull(buf, methods); err != nil {
		log.Error().Err(err).Msg("failed to read methods")
		return -1
	}

	if !bytes.Contains(methods, []byte{AuthUsernamePass}) {
		writeStatus(conn, RepGeneralFailure)
		log.Error().Msg("socks5: auth required")
		return -1
	}

	// Select username/password auth
	conn.Write([]byte{Version5, AuthUsernamePass})

	username, password, err := parseSocksAuth(buf)
	if err != nil {
		conn.Write([]byte{0x01, 0x01}) // auth version, failure
		log.Error().Err(err).Msg("failed to parse auth")
		return -1
	}

	username, paramStr := auth.SplitParams(username)
	if !auth.Verify(username, password) {
		conn.Write([]byte{0x01, 0x01})
		log.Error().Msg("failed to verify auth")
		return -1
	}
	params := auth.GetParams(paramStr)
	conn.Write([]byte{0x01, 0x00})

	hdr := make([]byte, 4)
	if _, err := io.ReadFull(buf, hdr); err != nil {
		log.Error().Err(err).Msg("failed to read header")
		return -1
	}

	addrType := hdr[3]
	ip, port, err := parseAtyp(addrType, buf)
	if err != nil {
		writeStatus(conn, RepAddrTypeNotSupported)
		log.Error().Err(err).Msg("failed to parse address")
		return -1
	}

	dialer, err := nio.GetDialer(
		ip,
		params[auth.ParamSession],
		params[auth.ParamTimeout],
		params[auth.ParamLocation],
		params[auth.ParamFallback],
	)
	if err != nil {
		writeStatus(conn, RepGeneralFailure)
		log.Error().Err(err).Msg("failed to get dialer")
		return -1
	}

	destConn, err := dialer.Dial("tcp", ip+":"+strconv.Itoa(int(port)))
	if err != nil {
		writeStatus(conn, RepHostUnreachable)
		log.Error().Err(err).Msg("failed to dial")
		return -1
	}
	defer destConn.Close()

	writeStatus(conn, RepSuccess)

	return nio.CopyOnce(destConn, conn, time.Duration(config.Get().MaxTimeout)*time.Second)
}

// writeStatus writes a standardized SOCKS5 reply to the client
func writeStatus(conn net.Conn, reply byte) {
	resp := []byte{
		Version5, reply, 0x00, AtypIPv4,
		0x00, 0x00, 0x00, 0x00, // Dummy IP
		0x00, 0x00, // Dummy port
	}
	conn.Write(resp)
}

// parseSocksAuth parses the username and password from the SOCKS5 authentication request
func parseSocksAuth(buf *bufio.Reader) (string, string, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(buf, header); err != nil {
		return "", "", fmt.Errorf("read auth header: %w", err)
	}

	ulen := int(header[1])
	ubuf := make([]byte, ulen)
	if _, err := io.ReadFull(buf, ubuf); err != nil {
		return "", "", fmt.Errorf("read username: %w", err)
	}

	if _, err := io.ReadFull(buf, header[:1]); err != nil {
		return "", "", fmt.Errorf("read password length: %w", err)
	}
	plen := int(header[0])
	pbuf := make([]byte, plen)
	if _, err := io.ReadFull(buf, pbuf); err != nil {
		return "", "", fmt.Errorf("read password: %w", err)
	}

	return string(ubuf), string(pbuf), nil
}

// parseAtyp parses the address type and returns the address and port
func parseAtyp(atyp byte, buf *bufio.Reader) (string, uint16, error) {
	switch atyp {
	case AtypIPv4:
		addr := make([]byte, 6)
		if _, err := io.ReadFull(buf, addr); err != nil {
			return "", 0, fmt.Errorf("read IPv4: %w", err)
		}
		return net.IP(addr[:4]).String(), binary.BigEndian.Uint16(addr[4:]), nil

	case AtypDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(buf, lenBuf); err != nil {
			return "", 0, fmt.Errorf("read domain length: %w", err)
		}
		domainLen := int(lenBuf[0])
		domainBuf := make([]byte, domainLen+2)
		if _, err := io.ReadFull(buf, domainBuf); err != nil {
			return "", 0, fmt.Errorf("read domain+port: %w", err)
		}

		host := string(domainBuf[:domainLen])
		port := binary.BigEndian.Uint16(domainBuf[domainLen:])
		ip, err := nio.ResolveHostname(host)
		if err != nil {
			return "", 0, fmt.Errorf("resolve hostname: %w", err)
		}
		return ip, port, nil

	case AtypIPv6:
		addr := make([]byte, 18)
		if _, err := io.ReadFull(buf, addr); err != nil {
			return "", 0, fmt.Errorf("read IPv6: %w", err)
		}
		ip := net.IP(addr[:16])
		port := binary.BigEndian.Uint16(addr[16:])
		return "[" + ip.String() + "]", port, nil

	default:
		return "", 0, fmt.Errorf("unsupported address type: %d", atyp)
	}
}
