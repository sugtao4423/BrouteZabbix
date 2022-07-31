package rl7023

import (
	"bytes"
	"strings"
	"time"

	"github.com/sugtao4423/BrouteZabbix/log"
	"go.bug.st/serial"
)

type Port struct {
	serial.Port
	lastBuff        []byte
	ReadTimeout     time.Duration
	ReadLineTimeout time.Duration
}

func NewPort(port serial.Port) *Port {
	return &Port{
		Port:     port,
		lastBuff: []byte{},
	}
}

func (p *Port) SetReadTimeout(d time.Duration) {
	p.ReadTimeout = d
}

func (p *Port) SetReadLineTimeout(d time.Duration) {
	p.ReadLineTimeout = d
}

func (p *Port) findLine() (bool, string) {
	nIndex := bytes.Index(p.lastBuff, []byte("\n"))
	if nIndex != -1 {
		line := string(p.lastBuff[:nIndex])
		line = strings.TrimSuffix(line, "\r")
		if len(p.lastBuff) > nIndex+1 {
			p.lastBuff = p.lastBuff[nIndex+1:]
		} else {
			p.lastBuff = []byte{}
		}
		return true, line
	}
	return false, ""
}

func (p *Port) ReadLine() string {
	found, line := p.findLine()
	if found {
		return line
	}

	err := p.Port.SetReadTimeout(p.ReadTimeout)
	if err != nil {
		log.Error("Error setting read timeout:", err)
	}

	deadline := time.Now().Add(p.ReadLineTimeout)
	buff := make([]byte, 100)
	for {
		if time.Now().After(deadline) {
			log.Warn("Timeout reading line")
			return ""
		}
		n, err := p.Port.Read(buff)
		if err != nil {
			log.Error("Error reading line:", err)
			continue
		}
		p.lastBuff = append(p.lastBuff, buff[:n]...)
		found, line = p.findLine()
		if found {
			return line
		}
	}
}
