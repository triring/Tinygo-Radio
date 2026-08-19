# Si4703

FMラジオ受信DSPモジュールSi4703をTinygoでコントロールするためのパッケージです。

<!--  ![./images/DSCN0954_800x600.jpg](./images/DSCN0954_800x600.jpg) -->

## Si4703の特徴

Skyworks（元Silicon Labs）のSi4703は、76MHz〜108MHzに対応するワンチップのFMラジオチューナーICです。RDS/RBDSデータ受信機能やデジタルボリューム調整を備え、Arduinoなどのマイコンと組み合わせてブレイクアウトボード（評価モジュール）でよく使われています。

* 受信周波数：76MHz 〜 108MHz（日本のFM放送やワイドFM、海外のFM放送に対応）
* RDS/RBDS（ラジオデータシステム）のデコード
* インターフェース：2線式 /3線式（I2C対応）

## 使用方法

以下のコマンドで、このリポジトリの内容をローカルにコピーして下さい。

```bash
git clone https://github.com/triring/Tinygo-Radio.git
```

コピーされたTinygo-Radio/si4703 ディレクトリ内のファイル構成
```bash
|   go.mod
|   go.sum
|   README.md
|   si4703.go
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
    +---MuteTest    無音(ミュート)、有音の切替え
    |       main.go
    |
    +---SI4703cmd  コマンドにより、Si4703を制御する。
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

1. 使用するターゲットボードとSi4703ラジオモジュールをI2Cケーブルで接続して下さい。
2. PCとターゲットボードをUSBケーブルで接続して下さい。
3. 最初に、このREADME.mdがあるディレクトリで、1度だけ以下のコマンドを実行して下さい。

```bash
go mod init si4703
go mod tidy
go: finding module for package tinygo.org/x/drivers
go: found tinygo.org/x/drivers in tinygo.org/x/drivers v0.35.0

go get github.com/triring/Tinygo-Radio/si4703

```
4. 実行したいコードのあるディレクトリ内に移動して下さい。その中にあるmain.goを開き、必要に応じて、お住まいの地域で、受信できるFM放送局の周波数をfreq変数に設定して下さい。(KHzで設定して下さい。)

```bash
	freq := 77700   // KHzで設定すること
```

5. 以下のコマンドで、コンパイル&実行ファイルの転送を行って下さい。今回は、Raspberrypi picoの互換ボードを使用しているので、-targetオプションをpicoと設定しています。他のマイコンボードを使用する場合は、そのボードに合わせて修正して下さい。

```bash
tinygo flash -target=pico -size=short -monitor .
```
