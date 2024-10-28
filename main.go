package main

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spetr/go-zabbix-sender"
	"github.com/sugtao4423/BrouteZabbix/log"
	"github.com/sugtao4423/BrouteZabbix/rl7023"
)

func main() {
	device := flag.String("device", "/dev/ttyUSB0", "serial device")
	brouteId := flag.String("bId", "", "B route id")
	broutePass := flag.String("bPass", "", "B route pass")
	checkInterval := flag.Int("interval", 60, "checking interval")
	zabbixServerHost := flag.String("zabbixServerHost", "localhost:10051", "Zabbix server host")
	zbxItemHostname := flag.String("zbxItemHostname", "", "Zabbix item hostname")
	zbxItemKey := flag.String("zbxItemKey", "", "Zabbix item key")
	flag.Parse()

	if strings.TrimSpace(*brouteId) == "" ||
		strings.TrimSpace(*broutePass) == "" ||
		strings.TrimSpace(*zbxItemHostname) == "" ||
		strings.TrimSpace(*zbxItemKey) == "" {
		flag.Usage()
		os.Exit(1)
	}

	rl7023 := rl7023.NewRL7023(*device)
	zbxSender := zabbix.NewSender(*zabbixServerHost)

	log.Info("Initializing RL7023")
	ipv6Addr := initRL7023(&rl7023, *brouteId, *broutePass)
	defer rl7023.Close()
	log.Info("RL7023 initialized")

	log.Info("Start checking")
	frame := []byte{0x10, 0x81, 0x00, 0x01, 0x05, 0xFF, 0x01, 0x02, 0x88, 0x01, 0x62, 0x01, 0xE7, 0x00}
	for {
		erxudp, err := rl7023.SKSENDTO("1", ipv6Addr, "0E1A", "1", frame)
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second * time.Duration(*checkInterval/4))
			continue
		}
		watt, err := getWatt(erxudp)
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second * time.Duration(*checkInterval/4))
			continue
		}
		log.Infof("瞬時電力計測値: %sW", watt)

		metrics := []*zabbix.Metric{
			zabbix.NewMetric(*zbxItemHostname, *zbxItemKey, watt, false),
		}
		_, _, res, err := zbxSender.SendMetrics(metrics)
		if err != nil {
			log.Error(err)
		}
		log.Debug(res)

		time.Sleep(time.Second * time.Duration(*checkInterval))
	}
}

// return ipv6 address
func initRL7023(device **rl7023.RL7023, brouteId string, broutePass string) string {
	rl7023 := *device

	// Connect
	log.Info("Connecting to RL7023")
	err := rl7023.Connect()
	if err != nil {
		log.Error("Error connecting to serial device.")
		log.Error("Exiting...")
		os.Exit(1)
	}
	log.Info("Connected to device")

	// SKRESET
	log.Info("Resetting device")
	err = rl7023.SKRESET()
	if err != nil {
		log.Error("Error resetting device.")
		log.Error("Exiting...")
		os.Exit(1)
	}
	log.Info("Device reset")

	// SKVER
	log.Info("Getting device version")
	err = rl7023.SKVER()
	if err != nil {
		log.Error("Error getting device version.")
		log.Error("Exiting...")
		os.Exit(1)
	}
	log.Info("Device version retrieved")

	// SKSETPWD
	log.Info("Setting B route password")
	err = rl7023.SKSETPWD(broutePass)
	if err != nil {
		log.Error("Error setting B route password.")
		log.Error("Exiting...")
		os.Exit(1)
	}
	log.Info("B route password set")

	// SKSETRBID
	log.Info("Setting B route ID")
	err = rl7023.SKSETRBID(brouteId)
	if err != nil {
		log.Error("Error setting B route ID.")
		log.Error("Exiting...")
		os.Exit(1)
	}
	log.Info("B route ID set")

	// SKSCAN
	log.Info("Scanning for PAN")
	pan, err := rl7023.SKSCAN()
	if err != nil {
		log.Error("Error scanning for PAN.")
		log.Error("Exiting...")
		os.Exit(1)
	}
	log.Info("PAN scanned")

	// SKSREG S2 Channel
	log.Info("Setting S2 Channel")
	err = rl7023.SKSREG("S2", pan.Channel)
	if err != nil {
		log.Error("Error setting S2.")
		log.Error("Exiting...")
		os.Exit(1)
	}
	log.Info("S2 Channel set")

	// SKSREG S3 PanId
	log.Info("Setting S3 PanId")
	err = rl7023.SKSREG("S3", pan.PanId)
	if err != nil {
		log.Error("Error setting S3.")
		log.Error("Exiting...")
		os.Exit(1)
	}
	log.Info("S3 PanId set")

	// SKLL64
	log.Info("Getting IPv6 address")
	ipv6Addr, err := rl7023.SKLL64(pan.Addr)
	if err != nil {
		log.Error("Error getting IPv6 address.")
		log.Error("Exiting...")
		os.Exit(1)
	}
	log.Info("IPv6 address retrieved")

	// SKJOIN
	log.Info("Start SKJOIN")
	err = rl7023.SKJOIN(ipv6Addr)
	if err != nil {
		log.Error("Error SKJOIN.")
		log.Error("Exiting...")
		os.Exit(1)
	}
	log.Info("SKJOIN finished")

	return ipv6Addr
}

func getWatt(erxudp string) (string, error) {
	cols := strings.Split(erxudp, " ")
	res := cols[8]
	seoj := res[8 : 8+6]
	esv := res[20 : 20+2]
	if seoj == "028801" && esv == "72" {
		epc := res[24 : 24+2]
		if epc == "E7" {
			watt, err := strconv.ParseUint(res[len(res)-8:], 16, 0)
			if err != nil {
				return "", err
			}
			return strconv.FormatUint(watt, 10), nil
		}
	}
	return "", errors.New("WARN Nothing watt")
}
