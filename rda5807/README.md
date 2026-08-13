# RDA5807

FMラジオ受信DSPモジュールRDA5807をTinygoでコントロールするためのパッケージです。

<!-- ![./images/DSCN0954_800x600.jpg](./images/DSCN0954_800x600.jpg) -->

## RDA5807の特徴

RDA5807シリーズ（主にRDA5807M）は、RDA Microelectronics製のFMラジオ受信DSPモジュールで、低消費電力、I²C制御、自動同調(AFC)、ステレオ復調を備えている。  
外付け部品が少なく、組み込み機器への搭載が容易である。さらに内蔵アンテナ入力の柔軟性や高感度受信が評価され、携帯端末やDIYラジオで広く利用されている。  

## 使用方法

以下のコマンドで、このリポジトリの内容をローカルにコピーして下さい。

```bash
git clone https://github.com/triring/Tinygo-Radio.git
```

コピーされたTinygo-Radio/rda5807 ディレクトリ内のファイル構成
```bash
|   go.mod
|   go.sum
|   README.md
|   rda5807.go
|
\---examples
    +---FMstation   配列に登録されている周波数を順番に受信していく。
    |       main.go
    |
    +---FreqScan    全周波数帯の信号強度を取得
    |       main.go
    |
    +---GetRSSI     信号強度を取得
    |       main.go
    |
    +---GetStatus   受信状態の取得
    |       main.go
    |
    +---MuteTest    無音(ミュート)、有音の切替え
    |       main.go
    |
    +---RDA5807cmd  コマンドにより、RDA5807を制御する。
    |       main.go
    |
    +---Simple      もっとも単純な受信テスト
    |       main.go
    |
    \---VolumeTest  音量変更のテスト
        main.go
```

コピーしたディレクトリ内に、examplesディレクトリがあります。  
この中にテスト用コードがあります。  

1. 使用するターゲットボードとRDA5807ラジオモジュールをI2Cケーブルで接続して下さい。
2. PCとターゲットボードをUSBケーブルで接続して下さい。
3. 最初に、このREADME.mdがあるディレクトリで、1度だけ以下のコマンドを実行して下さい。

```bash
go mod init rda5807
go mod tidy
go: finding module for package tinygo.org/x/drivers
go: found tinygo.org/x/drivers in tinygo.org/x/drivers v0.35.0

go get github.com/triring/Tinygo-Radio/rda5807

```
4. 実行したいコードのあるディレクトリ内に移動して下さい。その中にあるmain.goを開き、必要に応じて、お住まいの地域で、受信できるFM放送局の周波数をfreq変数に設定して下さい。(KHzで設定して下さい。)

```bash
	freq := 77700   // KHzで設定すること
```

5. 以下のコマンドで、コンパイル&実行ファイルの転送を行って下さい。今回は、Raspberrypi picoの互換ボードを使用しているので、-targetオプションをpicoと設定しています。他のマイコンボードを使用する場合は、そのボードに合わせて修正して下さい。

```bash
tinygo flash -target=pico -size=short -monitor .
```
