# BrouteZabbix
Bルートから電力使用量を取得してZabbixへ送信

## Usage
option | description | default
--- | --- | ---
`-device` | RL7023 device | `/dev/ttyUSB0`
`-bId` | BルートID |
`-bPass` | Bルートパスワード |
`-checkInterval` | 定期チェック間隔(秒) | `60`
`-zabbixSenderPath` | ZabbixSenderパス | `zabbix_sender`
`-zabbixServerHost` | ZabbixServerホスト |
`-zabbixServerPort` | ZabbixServerホストポート | `10051`
`-zbxItemHostname` | ZabbixHostname |
`-zbxItemKey` | ZabbixHostKey |

## Build
For Raspberry PI 3

```
GOOS=linux GOARCH=arm GOARM=7 go build
```

## Refs
* [スマートメーターの情報を最安ハードウェアで引っこ抜く - Qiita](https://qiita.com/rukihena/items/82266ed3a43e4b652adb)
* [higebu/wattmonitor](https://github.com/higebu/wattmonitor)
