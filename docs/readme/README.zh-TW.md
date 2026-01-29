# D2R Multiablo - Diablo 2 Resurrected 多開工具

[English](README.md) | [繁體中文](docs/readme/README.zh-TW.md) | [简体中文](docs/readme/README.zh-CN.md) | [日本語](docs/readme/README.ja.md)

一款為 Windows 設計的 D2R (Diablo II: Resurrected) 多開輔助工具，使用 Go 語言編寫。

## 下載

你可以從 [Releases](https://github.com/chenwei791129/multiablo/releases) 頁面下載 Multiablo 的最新版本，或者使用以下直接連結：

[🚀 下載最新版本 multiablo.exe](https://github.com/chenwei791129/multiablo/releases/latest/download/multiablo.exe)

## 概述

Multiablo 透過持續監視並關閉 D2R 用於阻止多開的單一事件觸發器 (Event Handle)，實現 Diablo II: Resurrected 的多開功能。只需將此工具運行於背景，即可透過 Battle.net 啟動器啟動多個 D2R 實例，無需額外操作。

## 運作原理

D2R 透過建立名為 `DiabloII Check For Other Instances` 的 Windows 事件觸發器 (Event Handle) 阻止多重實例啟動。當 D2R 啟動時，會檢查該觸發器是否存在，若存在則拒絕啟動遊戲。

Multiablo 的操作方式如下：
1. 持續監視運行中的 D2R.exe 程序
2. 自動檢測並關閉 `DiabloII Check For Other Instances` 事件觸發器
3. 讓你隨時從 Battle.net 啟動器啟動多個 D2R 實例
4. 監視 `Agent.exe` 程序，並僅在運行 7 秒後終止，以最大化 Battle.net 啟動器的可用性

## 使用方式

### 基本使用方法

1. **執行 multiablo.exe**
2. **從 Battle.net 啟動器啟動 D2R**
3. **從其他 Battle.net 啟動器啟動額外的 D2R 實例！**

程序啟動後會自動開始監視。你可以透過 GUI 檢視被檢測的程序與事件觸發器的操作狀態。

### 防毒軟體誤判

部分防毒軟體可能會因 Multiablo 操作進程句柄而將其標記為威脅。這是此類工具的正常行為，你可能需要將其加入例外清單。

## 免責聲明

此工具僅供教育與個人用途使用，請自行承擔使用風險。作者對以下事項不承擔責任：
- 與 Diablo II: Resurrected 服務條款的任何違規行為
- 帳號的停權或封禁
- 遊戲崩潰或數據丟失
- 使用此軟件引發的任何其他問題

**注意**：多開可能違反該遊戲的服務條款。使用前請檢查 Blizzard 的政策。

## 授權

MIT 授權 - 詳情請見 LICENSE 檔案

## 致謝

- 靈感來源於 Process Explorer 的句柄管理功能
