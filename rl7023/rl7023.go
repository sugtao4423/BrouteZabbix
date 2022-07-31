package rl7023

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/sugtao4423/BrouteZabbix/log"
	"go.bug.st/serial"
)

type RL7023 struct {
	Baudrate     int
	SerialDevice string
	Port         serial.Port
	lastBuff     []byte
	timeout      time.Duration
}

func NewRL7023(device string) *RL7023 {
	return &RL7023{
		Baudrate:     115200,
		SerialDevice: device,
		lastBuff:     []byte{},
		timeout:      time.Duration(30) * time.Second,
	}
}

func (rl7023 *RL7023) setTimeout(sec int) {
	rl7023.timeout = time.Duration(sec) * time.Second
}

func (rl7023 *RL7023) readLine() string {
	findLine := func() (bool, string) {
		nIndex := bytes.Index(rl7023.lastBuff, []byte("\n"))
		if nIndex != -1 {
			line := string(rl7023.lastBuff[:nIndex])
			line = strings.TrimSuffix(line, "\r")
			if len(rl7023.lastBuff) > nIndex+1 {
				rl7023.lastBuff = rl7023.lastBuff[nIndex+1:]
			} else {
				rl7023.lastBuff = []byte{}
			}
			return true, line
		}
		return false, ""
	}

	found, line := findLine()
	if found {
		return line
	}

	err := rl7023.Port.SetReadTimeout(time.Duration(2) * time.Second)
	if err != nil {
		log.Error("Error setting read timeout:", err)
	}

	deadline := time.Now().Add(rl7023.timeout)
	buff := make([]byte, 100)
	for {
		if time.Now().After(deadline) {
			log.Warn("Timeout reading line")
			return ""
		}
		n, err := rl7023.Port.Read(buff)
		if err != nil {
			log.Error("Error reading line:", err)
			continue
		}
		rl7023.lastBuff = append(rl7023.lastBuff, buff[:n]...)
		found, line = findLine()
		if found {
			return line
		}
	}
}

func (rl7023 *RL7023) Connect() error {
	mode := &serial.Mode{
		BaudRate: rl7023.Baudrate,
	}
	port, err := serial.Open(rl7023.SerialDevice, mode)
	if err != nil {
		return err
	}
	rl7023.Port = port
	rl7023.setTimeout(30)
	return nil
}

func (rl7023 *RL7023) write(s string) error {
	_, err := rl7023.Port.Write([]byte(s))
	if err != nil {
		return err
	}
	return nil
}

// includes `log.Debug`
func (rl7023 *RL7023) readLinesUntilOK() []string {
	var lines []string
	for {
		line := rl7023.readLine()
		log.Debug(line)
		lines = append(lines, line)
		if line == "OK" {
			return lines
		}
	}
}

func (rl7023 *RL7023) Close() error {
	return rl7023.Port.Close()
}

func (rl7023 *RL7023) SKVER() error {
	err := rl7023.write("SKVER\r\n")
	if err != nil {
		return err
	}
	rl7023.readLinesUntilOK()
	return nil
}

func (rl7023 *RL7023) SKSETPWD(broutePass string) error {
	err := rl7023.write("SKSETPWD C " + broutePass + "\r\n")
	if err != nil {
		return err
	}
	rl7023.readLinesUntilOK()
	return nil
}

func (rl7023 *RL7023) SKSETRBID(brouteId string) error {
	err := rl7023.write("SKSETRBID " + brouteId + "\r\n")
	if err != nil {
		return err
	}
	rl7023.readLinesUntilOK()
	return nil
}

func (rl7023 *RL7023) SKSCAN() (*PAN, error) {
	err := rl7023.write("SKSCAN 2 FFFFFFFF 6\r\n")
	if err != nil {
		return nil, err
	}

	pan := &PAN{}
	for {
		line := rl7023.readLine()
		log.Debug(line)
		s := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(s, "Channel:"):
			pan.Channel = strings.Split(s, ":")[1]
		case strings.HasPrefix(s, "Channel Page:"):
			pan.ChannelPage = strings.Split(s, ":")[1]
		case strings.HasPrefix(s, "Pan ID:"):
			pan.PanId = strings.Split(s, ":")[1]
		case strings.HasPrefix(s, "Addr:"):
			pan.Addr = strings.Split(s, ":")[1]
		case strings.HasPrefix(s, "LQI:"):
			pan.LQI = strings.Split(s, ":")[1]
		case strings.HasPrefix(s, "PairID:"):
			pan.PairId = strings.Split(s, ":")[1]
		}
		if strings.HasPrefix(s, "EVENT 22") {
			break
		}
	}
	if pan.Channel == "" ||
		pan.ChannelPage == "" ||
		pan.PanId == "" ||
		pan.Addr == "" ||
		pan.LQI == "" ||
		pan.PairId == "" {
		return nil, fmt.Errorf("SKSCAN failed")
	}
	return pan, nil
}

func (rl7023 *RL7023) SKSREG(key string, val string) error {
	err := rl7023.write("SKSREG " + key + " " + val + "\r\n")
	if err != nil {
		return err
	}
	rl7023.readLinesUntilOK()
	return nil
}

func (rl7023 *RL7023) SKLL64(addr string) (string, error) {
	err := rl7023.write("SKLL64 " + addr + "\r\n")
	if err != nil {
		return "", err
	}

	var lines []string
	for i := 0; i < 2; i++ {
		line := rl7023.readLine()
		log.Debug(string(line))
		lines = append(lines, string(line))
	}
	return lines[1], nil
}

func (rl7023 *RL7023) SKJOIN(ipv6Addr string) error {
	err := rl7023.write("SKJOIN " + ipv6Addr + "\r\n")
	if err != nil {
		return err
	}
	rl7023.readLinesUntilOK()

	for {
		line := rl7023.readLine()
		log.Debug(line)
		if strings.HasPrefix(line, "EVENT 24") {
			return fmt.Errorf("SKJOIN failed. %s", line)
		} else if strings.HasPrefix(line, "EVENT 25") {
			break
		}
	}

	rl7023.setTimeout(2)
	line := rl7023.readLine()
	log.Debug(line)
	return nil
}

func (rl7023 *RL7023) SKSENDTO(handle string, ipAddr string, port string, sec string, data []byte) (string, error) {
	rl7023.setTimeout(2)

	base := fmt.Sprintf("SKSENDTO %s %s %s %s %.4X ", handle, ipAddr, port, sec, len(data))
	cmd := append([]byte(base), data[:]...)
	cmd = append(cmd, []byte("\r\n")[:]...)
	_, err := rl7023.Port.Write(cmd)
	if err != nil {
		return "", err
	}

	var lines []string
	for i := 0; i < 5; i++ {
		line := rl7023.readLine()
		log.Debug(string(line))
		lines = append(lines, string(line))
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "ERXUDP") {
			return line, nil
		}
	}
	return "", fmt.Errorf("ERXUDP Nothing. %s", lines)
}
