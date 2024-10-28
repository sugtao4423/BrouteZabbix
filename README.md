# BrouteZabbix
Bルートから電力使用量を取得してZabbixへ送信

## Usage
option | description | default
--- | --- | ---
`-device` | RL7023 device | `/dev/ttyUSB0`
`-bId` | BルートID |
`-bPass` | Bルートパスワード |
`-checkInterval` | 定期チェック間隔(秒) | `60`
`-zabbixServerHost` | ZabbixServerホスト | `localhost:10051`
`-zbxItemHostname` | ZabbixHostname |
`-zbxItemKey` | ZabbixHostKey |

## Build
For Raspberry Pi 3

```
GOOS=linux GOARCH=arm GOARM=7 go build
```

## systemd
```
# /etc/systemd/system/BrouteZabbix.service
[Unit]
After=network.target

[Service]
Type=simple
ExecStart=/path/to/BrouteZabbix \
            -bId BROUTE_ID \
            -bPass BROUTE_PASS \
            -zabbixServerHost ZABBIX_SERVER:10051 \
            -zbxItemHostname ZABBIX_ITEM_HOSTNAME \
            -zbxItemKey ZABBIX_ITEM_KEY
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Refs
* [スマートメーターの情報を最安ハードウェアで引っこ抜く - Qiita](https://qiita.com/rukihena/items/82266ed3a43e4b652adb)
* [higebu/wattmonitor](https://github.com/higebu/wattmonitor)
