# BrouteZabbix
`RL7023 Stick-D/IPS`を使ってBルートで電力使用量を取得  
そしてZabbixへ送信

Pythonスクリプトを使わせていただきました！  
[スマートメーターの情報を最安ハードウェアで引っこ抜く - Qiita](https://qiita.com/rukihena/items/82266ed3a43e4b652adb)

## conf
そんなものはない  
スクリプト書き換えてね

* `broute.py`
    - `rbid`
    - `rbpwd`

## script.sh
#### start
```
./script start
```

#### stop
```
./script stop
```

#### send zabbix
```
./script send
```
