package zabbix_sender

import (
	"fmt"
	"os/exec"
	"strings"
)

type ZabbixSender struct {
	ZabbixSenderPath string
	ZabbixHost       string
	ZabbixPort       int
}

func NewZabbixSender(zabbixSenderPath string, zabbixHost string, zabbixPort int) *ZabbixSender {
	return &ZabbixSender{
		ZabbixSenderPath: zabbixSenderPath,
		ZabbixHost:       zabbixHost,
		ZabbixPort:       zabbixPort,
	}
}

func (zabbixSender *ZabbixSender) Send(hostname string, key string, value string) (string, error) {
	command := fmt.Sprintf(
		"%s -z %s -p %d -s %s -k %s -o %s | head -1",
		zabbixSender.ZabbixSenderPath,
		zabbixSender.ZabbixHost,
		zabbixSender.ZabbixPort,
		hostname,
		key,
		value,
	)

	out, err := exec.Command("sh", "-c", command).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
