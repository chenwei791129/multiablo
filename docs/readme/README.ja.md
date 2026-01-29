# D2R Multiablo - Diablo 2 Resurrected マルチインスタンスツール

[English](README.md) | [繁體中文](docs/readme/README.zh-TW.md) | [简体中文](docs/readme/README.zh-CN.md) | [日本語](docs/readme/README.ja.md)

Windows向けにGoで作成された、D2R（Diablo II: Resurrected）のマルチインスタンスヘルパーツールです。

## ダウンロード

Multiabloの最新版は[Releases](https://github.com/chenwei791129/multiablo/releases)ページから、または以下の直接リンクからダウンロードできます。

[🚀 最新版 multiablo.exe をダウンロードする](https://github.com/chenwei791129/multiablo/releases/latest/download/multiablo.exe)

## 概要

Multiabloは、D2Rがマルチ起動を防ぐために使用する単一インスタンスのイベントハンドルを継続的に監視して閉じることで、複数のDiablo II: Resurrectedインスタンスを同時に実行できるようにします。このツールをバックグラウンドで実行するだけで、Battle.netランチャーから追加の操作なく、複数のD2Rインスタンスを起動できます。

## 動作の仕組み

D2Rは、`DiabloII Check For Other Instances`という名前のWindowsイベントハンドルを作成することで、複数インスタンスの起動を防ぎます。D2Rが起動すると、このハンドルが存在するか確認し、存在するとゲームの起動を拒否します。

Multiabloは以下のように動作します：
1. 実行中のD2R.exeプロセスを継続的に監視します
2. 自動的に`DiabloII Check For Other Instances`イベントハンドルを検出して閉じます
3. Battle.netランチャーからいつでも複数のD2Rインスタンスを起動可能にします
4. `Agent.exe`プロセスを監視し、起動後7秒経過してから終了させることで、Battle.netランチャーでのゲーム起動可能性を最大化します

## 使用方法

### 基本的な使い方

1. **multiablo.exeを実行する**
2. Battle.netランチャーから **D2Rを起動する**
3. 他のBattle.netランチャーから **追加のD2Rインスタンスを起動する**

ツールを起動すると、自動的に監視が開始されます。GUIで検出されたプロセスやハンドル操作の状況を確認できます。

### ウイルス対策ソフトによる誤検出

一部のウイルス対策ソフトウェアでは、Multiabloがプロセスハンドルを操作するため、警告が出る場合があります。このツールの特性上、これは想定される動作です。必要に応じて例外設定を行ってください。

## 免責事項

このツールは教育目的および個人利用のみを目的としています。使用は自己責任で行ってください。作者は以下について責任を負いません：
- Diablo II: Resurrected利用規約の違反
- アカウントの停止やBAN
- ゲームのクラッシュやデータ損失
- このソフトウェアの使用に起因するその他の問題

**注意**：複数インスタンスの実行はゲームの利用規約に反する場合があります。使用前にBlizzardのポリシーを確認してください。

## ライセンス

MITライセンス - 詳細はLICENSEファイルをご参照ください

## 謝辞

- Process Explorerのハンドル管理機能に触発されました
