package util

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

const (
	PREFIX_UNIX = "unix://"
	PREFIX_TCP  = "tcp://"
)

type simpleAddr struct {
	network string
	addr    string
}

func (sa simpleAddr) Network() string {
	return sa.network
}

func (sa simpleAddr) String() string {
	return sa.addr
}

func Parse(u string) (net.Addr, error) {
	if len(u) < len(PREFIX_UNIX) && len(u) < len(PREFIX_TCP) {
		return nil, errors.New("url too short")
	}
	sa := new(simpleAddr)
	if strings.HasPrefix(u, PREFIX_UNIX) {
		sa.network = "unix"
		sa.addr = strings.TrimPrefix(u, PREFIX_UNIX)
	} else if strings.HasPrefix(u, PREFIX_TCP) {
		sa.network = "tcp"
		sa.addr = strings.TrimPrefix(u, PREFIX_TCP)
	} else {
		return nil, fmt.Errorf("unknown protocol %s", u)
	}

	return *sa, nil
}
