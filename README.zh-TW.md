# Multiablo

適用於 Windows 的《暗黑破壞神 II：獄火重生》(D2R) 多開輔助工具，使用 Go 語言開發。

## 概述

Multiablo 讓您能夠同時運行多個《暗黑破壞神 II：獄火重生》實例，透過持續監控並關閉 D2R 用來防止多開的單一實例事件控制碼。只需在背景執行此工具，您就可以從 Battle.net 啟動器啟動多個 D2R 實例，無需任何額外步驟。

## 工作原理

D2R 透過建立名為 `DiabloII Check For Other Instances` 的 Windows 事件控制碼來防止多開。當 D2R 啟動時，它會檢查此控制碼是否存在 - 如果存在，遊戲就會拒絕啟動。

Multiablo 的工作方式：
1. 持續監控正在運行的 D2R.exe 進程
2. 自動偵測並關閉 `DiabloII Check For Other Instances` 事件控制碼
3. 讓您可以隨時從 Battle.net 啟動器啟動多個 D2R 實例
4. 持續監控並終止可能干擾多開運作的 `Agent.exe` 進程

## 使用說明

### 基本用法

1. **從 Battle.net 啟動器啟動 D2R**
2. **執行 multiablo.exe**
3. **從其他 Battle.net 啟動器啟動額外的 D2R 實例**！

### 命令列選項

```
> multiablo.exe -h
Multiablo enables you to run multiple instances of Diablo II: Resurrected
simultaneously by continuously monitoring and removing the "DiabloII Check For Other Instances" and "Agent.exe".

Usage:
  multiablo [flags]

Flags:
  -h, --help      help for multiablo
  -v, --verbose   Enable verbose output
```

### 輸出範例

```
2025-12-21T00:11:09.963+0800    INFO    Multiablo - D2R Multi-Instance Helper
2025-12-21T00:11:09.983+0800    INFO    ======================================
2025-12-21T00:11:09.983+0800    INFO
2025-12-21T00:11:09.984+0800    INFO    Starting background monitors...
2025-12-21T00:11:09.984+0800    INFO    Monitoring Agent.exe processes for termination...
2025-12-21T00:11:09.984+0800    INFO    Monitoring D2R.exe processes for handle restrictions...
2025-12-21T00:11:09.994+0800    INFO
2025-12-21T00:11:09.994+0800    INFO    Press Enter to exit...
```

### 防毒軟體誤報

某些防毒軟體可能會將 Multiablo 標記為威脅，因為它會操作進程控制碼。這是此類工具的預期行為。您可能需要新增例外。

## 免責聲明

此工具僅供教育和個人使用。使用風險自負。作者不對以下情況負責：
- 任何違反《暗黑破壞神 II：獄火重生》服務條款的行為
- 帳號停權或封鎖
- 遊戲崩潰或資料遺失
- 因使用本軟體而產生的任何其他問題

**注意**：運行多個實例可能違反遊戲的服務條款。使用前請查看暴雪的政策。

## 授權

MIT 授權 - 詳見 LICENSE 檔案

## 致謝

- 靈感來自 Process Explorer 的控制碼管理功能
