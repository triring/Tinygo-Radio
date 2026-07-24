# TEA5767

FMラジオ受信DSPモジュールTEA5767をTinygoでコントロールするためのパッケージです。

## TEA5767の特徴

TEA5767は、NXP Semiconductors（旧Philips）製のFMラジオ受信DSPモジュールで、低消費電力・小型設計、I²C制御、自動同調(AFC)、ステレオ復調を備えている。  
外付け部品が少なく、組み込み機器や携帯端末への実装が容易で、安定した受信性能を手軽に利用できる。  

## 使用方法

以下のコマンドで、このリポジトリの内容をローカルにコピーして下さい。

```bash
git clone https://github.com/triring/tm1638.git
```

コピーされたtm1638ディレクトリ内のファイル構成
```bash
D:.
|   .gitignore


コピーしたディレクトリ内に、examplesディレクトリがあります。
この中にテスト用コードがあります。

1. ターゲットボードとtm1638評価ボードを3本の信号線、電源、GND線で接続して下さい。
2. PCとターゲットボードをUSBケーブルで接続して下さい。
3. 実行したいコードのあるディレクトリ内に移動して下さい。
4. 最初に、1度だけ以下のコマンドを実行して下さい。

```bash
go mod init tea5767
go mod tidy
go: finding module for package tinygo.org/x/drivers
go: found tinygo.org/x/drivers in tinygo.org/x/drivers v0.35.0
```
5. 以下のコマンドで、コンパイル&実行ファイルの転送を行って下さい。  
(-targetオプションは、使用するマイコンボードに合わせて修正して下さい。)

```bash
tinygo flash -target=pico -size=short -monitor .
```




## 問題点
このチップが設計されたのは、
ワイドFM（90.0MHz以上）の放送が始まる前である。
よって、受信周波数帯を日本エリアに設定するとワイドFMは受信できない。
そこで、受信周波数帯をUS/Euroエリアに切り替えて受信することになる。

この状態に陥ると、


によって、
ワイドFM（90.0MHz以上）を受信するためのバンド変更