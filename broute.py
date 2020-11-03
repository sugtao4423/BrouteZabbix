#!/usr/bin/env python
# -*- coding: utf-8 -*-

#
# Thanks!!!
#   スマートメーターの情報を最安ハードウェアで引っこ抜く - Qiita
#   https://qiita.com/rukihena/items/82266ed3a43e4b652adb
#

from __future__ import print_function

import sys
import serial
import time

# Bルート認証ID（東京電力パワーグリッドから郵送で送られてくるヤツ）
rbid  = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
# Bルート認証パスワード（東京電力パワーグリッドからメールで送られてくるヤツ）
rbpwd = "XXXXXXXXXXXX"
# シリアルポートデバイス名
#serialPortDev = 'COM3'  # Windows の場合
serialPortDev = '/dev/ttyUSB0'  # Linux(ラズパイなど）の場合
#serialPortDev = '/dev/cu.usbserial-A103BTPR'    # Mac の場合

# シリアルポート初期化
ser = serial.Serial(serialPortDev, 115200)

# とりあえずバージョンを取得してみる（やらなくてもおｋ）
ser.write("SKVER\r\n")
print(ser.readline(), end="") # エコーバック
print(ser.readline(), end="") # バージョン

# Bルート認証パスワード設定
ser.write("SKSETPWD C " + rbpwd + "\r\n")
print(ser.readline(), end="") # エコーバック
print(ser.readline(), end="") # OKが来るはず（チェック無し）

# Bルート認証ID設定
ser.write("SKSETRBID " + rbid + "\r\n")
print(ser.readline(), end="") # エコーバック
print(ser.readline(), end="") # OKが来るはず（チェック無し）

scanDuration = 4;   # スキャン時間。サンプルでは6なんだけど、4でも行けるので。（ダメなら増やして再試行）
scanRes = {} # スキャン結果の入れ物

# スキャンのリトライループ（何か見つかるまで）
while not scanRes.has_key("Channel") :
    # アクティブスキャン（IE あり）を行う
    # 時間かかります。10秒ぐらい？
    ser.write("SKSCAN 2 FFFFFFFF " + str(scanDuration) + "\r\n")

    # スキャン1回について、スキャン終了までのループ
    scanEnd = False
    while not scanEnd :
        line = ser.readline()
        print(line, end="")

        if line.startswith("EVENT 22") :
            # スキャン終わったよ（見つかったかどうかは関係なく）
            scanEnd = True
        elif line.startswith("  ") :
            # スキャンして見つかったらスペース2個あけてデータがやってくる
            # 例
            #  Channel:39
            #  Channel Page:09
            #  Pan ID:FFFF
            #  Addr:FFFFFFFFFFFFFFFF
            #  LQI:A7
            #  PairID:FFFFFFFF
            cols = line.strip().split(':')
            scanRes[cols[0]] = cols[1]
    scanDuration+=1

    if 7 < scanDuration and not scanRes.has_key("Channel"):
        # 引数としては14まで指定できるが、7で失敗したらそれ以上は無駄っぽい
        print("スキャンリトライオーバー")
        sys.exit()  #### 糸冬了 ####

# スキャン結果からChannelを設定。
ser.write("SKSREG S2 " + scanRes["Channel"] + "\r\n")
print(ser.readline(), end="") # エコーバック
print(ser.readline(), end="") # OKが来るはず（チェック無し）

# スキャン結果からPan IDを設定
ser.write("SKSREG S3 " + scanRes["Pan ID"] + "\r\n")
print(ser.readline(), end="") # エコーバック
print(ser.readline(), end="") # OKが来るはず（チェック無し）

# MACアドレス(64bit)をIPV6リンクローカルアドレスに変換。
# (BP35A1の機能を使って変換しているけど、単に文字列変換すればいいのではという話も？？)
ser.write("SKLL64 " + scanRes["Addr"] + "\r\n")
print(ser.readline(), end="") # エコーバック
ipv6Addr = ser.readline().strip()
print(ipv6Addr)

# PANA 接続シーケンスを開始します。
ser.write("SKJOIN " + ipv6Addr + "\r\n");
print(ser.readline(), end="") # エコーバック
print(ser.readline(), end="") # OKが来るはず（チェック無し）

# PANA 接続完了待ち（10行ぐらいなんか返してくる）
bConnected = False
while not bConnected :
    line = ser.readline()
    print(line, end="")
    if line.startswith("EVENT 24") :
        print("PANA 接続失敗")
        sys.exit()  #### 糸冬了 ####
    elif line.startswith("EVENT 25") :
        # 接続完了！
        bConnected = True

# これ以降、シリアル通信のタイムアウトを設定
ser.timeout = 2

# スマートメーターがインスタンスリスト通知を投げてくる
# (ECHONET-Lite_Ver.1.12_02.pdf p.4-16)
print(ser.readline(), end="") #無視

while True:
    # ECHONET Lite フレーム作成
    # 　参考資料
    # 　・ECHONET-Lite_Ver.1.12_02.pdf (以下 EL)
    # 　・Appendix_H.pdf (以下 AppH)
    echonetLiteFrame = ""
    echonetLiteFrame += "\x10\x81"      # EHD (参考:EL p.3-2)
    echonetLiteFrame += "\x00\x01"      # TID (参考:EL p.3-3)
    # ここから EDATA
    echonetLiteFrame += "\x05\xFF\x01"  # SEOJ (参考:EL p.3-3 AppH p.3-408～)
    echonetLiteFrame += "\x02\x88\x01"  # DEOJ (参考:EL p.3-3 AppH p.3-274～)
    echonetLiteFrame += "\x62"          # ESV(62:プロパティ値読み出し要求) (参考:EL p.3-5)
    echonetLiteFrame += "\x01"          # OPC(1個)(参考:EL p.3-7)
    echonetLiteFrame += "\xE7"          # EPC(参考:EL p.3-7 AppH p.3-275)
    echonetLiteFrame += "\x00"          # PDC(参考:EL p.3-9)

    # コマンド送信
    command = "SKSENDTO 1 {0} 0E1A 1 {1:04X} {2}".format(ipv6Addr, len(echonetLiteFrame), echonetLiteFrame)
    ser.write(command)

    print(ser.readline(), end="") # エコーバック
    print(ser.readline(), end="") # EVENT 21 が来るはず（チェック無し）
    print(ser.readline(), end="") # OKが来るはず（チェック無し）
    line = ser.readline()         # ERXUDPが来るはず
    print(line, end="")

    # 受信データはたまに違うデータが来たり、
    # 取りこぼしたりして変なデータを拾うことがあるので
    # チェックを厳しめにしてます。
    if line.startswith("ERXUDP") :
        cols = line.strip().split(' ')
        res = cols[8]   # UDP受信データ部分
        #tid = res[4:4+4];
        seoj = res[8:8+6]
        #deoj = res[14,14+6]
        ESV = res[20:20+2]
        #OPC = res[22,22+2]
        if seoj == "028801" and ESV == "72" :
            # スマートメーター(028801)から来た応答(72)なら
            EPC = res[24:24+2]
            if EPC == "E7" :
                # 内容が瞬時電力計測値(E7)だったら
                hexPower = line[-8:]    # 最後の4バイト（16進数で8文字）が瞬時電力計測値
                intPower = int(hexPower, 16)
                print(u"瞬時電力計測値:{0}[W]".format(intPower))



# 無限ループだからここには来ないけどな
ser.close()
