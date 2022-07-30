package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sugtao4423/BrouteZabbix/rl7023"
	"github.com/sugtao4423/BrouteZabbix/zabbix_sender"
)

func main() {
	device := flag.String("device", "/dev/ttyUSB0", "serial device")
	brouteId := flag.String("bId", "", "B route id")
	broutePass := flag.String("bPass", "", "B route pass")
	checkInterval := flag.Int("interval", 60, "checking interval")
	zabbixSenderPath := flag.String("zabbixSenderPath", "zabbix_sender", "zabbix sender path")
	zabbixServerHost := flag.String("zabbixServerHost", "", "Zabbix server host")
	zabbixServerPort := flag.Int("zabbixServerPort", 10051, "Zabbix server port")
	zbxItemHostname := flag.String("zbxItemHostname", "", "Zabbix item hostname")
	zbxItemKey := flag.String("zbxItemKey", "", "Zabbix item key")
	flag.Parse()

	if strings.TrimSpace(*brouteId) == "" ||
		strings.TrimSpace(*broutePass) == "" ||
		strings.TrimSpace(*zabbixSenderPath) == "" ||
		strings.TrimSpace(*zabbixServerHost) == "" ||
		strings.TrimSpace(*zbxItemHostname) == "" ||
		strings.TrimSpace(*zbxItemKey) == "" {
		flag.Usage()
		os.Exit(1)
	}

	rl7023 := rl7023.NewRL7023(*device)
	zbxSender := zabbix_sender.NewZabbixSender(*zabbixSenderPath, *zabbixServerHost, *zabbixServerPort)

	log.Println("> Initializing RL7023")
	ipv6Addr := initRL7023(&rl7023, *brouteId, *broutePass)
	defer rl7023.Close()
	log.Println("> RL7023 initialized")

	log.Println("> Start checking")
	frame := []byte{0x10, 0x81, 0x00, 0x01, 0x05, 0xFF, 0x01, 0x02, 0x88, 0x01, 0x62, 0x01, 0xE7, 0x00}
	for {
		erxudp, err := rl7023.SKSENDTO("1", ipv6Addr, "0E1A", "1", frame)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * time.Duration(*checkInterval/4))
			continue
		}
		watt, err := getWatt(erxudp)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * time.Duration(*checkInterval/4))
			continue
		}
		log.Printf("> 瞬時電力計測値: %sW", watt)

		res, err := zbxSender.Send(*zbxItemHostname, *zbxItemKey, watt)
		if err != nil {
			log.Println(err)
		}
		log.Println(res)

		time.Sleep(time.Second * time.Duration(*checkInterval))
	}
}

// return ipv6 address
func initRL7023(device **rl7023.RL7023, brouteId string, broutePass string) string {
	rl7023 := *device

	// Connect
	log.Println("> Connecting to RL7023")
	err := rl7023.Connect()
	if err != nil {
		log.Fatalln("Error connecting to serial device.")
		log.Fatalln("Exiting...")
		os.Exit(1)
	}
	log.Println("> Connected to device")

	// SKVER
	log.Println("> Getting device version")
	err = rl7023.SKVER()
	if err != nil {
		log.Fatalln("Error getting device version.")
		log.Fatalln("Exiting...")
		os.Exit(1)
	}
	log.Println("> Device version retrieved")

	// SKSETPWD
	log.Println("> Setting B route password")
	err = rl7023.SKSETPWD(broutePass)
	if err != nil {
		log.Fatalln("Error setting B route password.")
		log.Fatalln("Exiting...")
		os.Exit(1)
	}
	log.Println("> B route password set")

	// SKSETRBID
	log.Println("> Setting B route ID")
	err = rl7023.SKSETRBID(brouteId)
	if err != nil {
		log.Fatalln("Error setting B route ID.")
		log.Fatalln("Exiting...")
		os.Exit(1)
	}
	log.Println("> B route ID set")

	// SKSCAN
	log.Println("> Scanning for PAN")
	pan, err := rl7023.SKSCAN()
	if err != nil {
		log.Fatalln("Error scanning for PAN.")
		log.Fatalln("Exiting...")
		os.Exit(1)
	}
	log.Println("> PAN scanned")

	// SKSREG S2 Channel
	log.Println("> Setting S2 Channel")
	err = rl7023.SKSREG("S2", pan.Channel)
	if err != nil {
		log.Fatalln("Error setting S2.")
		log.Fatalln("Exiting...")
		os.Exit(1)
	}
	log.Println("> S2 Channel set")

	// SKSREG S3 PanId
	log.Println("> Setting S3 PanId")
	err = rl7023.SKSREG("S3", pan.PanId)
	if err != nil {
		log.Fatalln("Error setting S3.")
		log.Fatalln("Exiting...")
		os.Exit(1)
	}
	log.Println("> S3 PanId set")

	// SKLL64
	log.Println("> Getting IPv6 address")
	ipv6Addr, err := rl7023.SKLL64(pan.Addr)
	if err != nil {
		log.Fatalln("Error getting IPv6 address.")
		log.Fatalln("Exiting...")
		os.Exit(1)
	}
	log.Println("> IPv6 address retrieved")

	// SKJOIN
	log.Println("> Start SKJOIN")
	err = rl7023.SKJOIN(ipv6Addr)
	if err != nil {
		log.Fatalln("Error SKJOIN.")
		log.Fatalln("Exiting...")
		os.Exit(1)
	}
	log.Println("> SKJOIN finished")

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
				log.Println(err)
			}
			return strconv.FormatUint(watt, 10), nil
		}
	}
	return "", errors.New("WARN Nothing watt")
}
