# D2R Multiablo - Diablo 2 重制版多实例工具

[English](README.md) | [繁體中文](docs/readme/README.zh-TW.md) | [简体中文](docs/readme/README.zh-CN.md) | [日本語](docs/readme/README.ja.md)

一个适用于 Windows 的 D2R（Diablo II: Resurrected）多实例辅助工具，使用 Go 语言编写。

## 下载

您可以从 [Releases](https://github.com/chenwei791129/multiablo/releases) 页面下载 Multiablo 的最新版本，或者使用以下直接链接：

[🚀 下载最新版本 multiablo.exe](https://github.com/chenwei791129/multiablo/releases/latest/download/multiablo.exe)

## 概览

Multiablo 通过持续监控并关闭 D2R 用于防止多实例启动的单实例事件句柄，使您能够同时运行多个 Diablo II: Resurrected 实例。只需在后台运行此工具，您便可以从 Battle.net 启动器启动多个 D2R 实例，而无需任何额外操作。

## 工作原理

D2R 通过创建一个名为 `DiabloII Check For Other Instances` 的 Windows 事件句柄来防止多实例运行。当 D2R 启动时，它会检查该句柄是否存在——如果存在，游戏将拒绝启动。

Multiablo 的工作方式如下：
1. 持续监控运行中的 D2R.exe 进程
2. 自动检测并关闭 `DiabloII Check For Other Instances` 事件句柄
3. 允许您随时从 Battle.net 启动器启动多个 D2R 实例
4. 监控 `Agent.exe` 进程，并仅在运行 7 秒后才终止它，从而最大化 Battle.net 启动器的可用性以启动游戏

## 使用方法

### 基本使用方法

1. **运行 multiablo.exe**
2. **从 Battle.net 启动器启动 D2R**
3. **通过其他 Battle.net 启动器启动额外的 D2R 实例！**

应用程序在启动时将自动开始监控。您可以在 GUI 中查看已检测到的进程和句柄操作的状态。

### 杀毒软件误报

某些杀毒软件可能会将 Multiablo 标记为威胁，因为它操作了进程句柄。这种行为是此类工具的预期行为。您可能需要添加例外。

## 免责声明

此工具仅供教育和个人使用。使用请自行承担风险。作者不对以下情况负责：
- 任何违反 Diablo II: Resurrected 服务条款的行为
- 因账户冻结或封禁导致的损失
- 游戏崩溃或数据丢失
- 其他因使用此软件产生的问题

**注意**：运行多个实例可能违反游戏的服务条款。在使用之前，请先查阅 Blizzard 的相关政策。

## 许可证

MIT 许可证 - 详见 LICENSE 文件

## 致谢

- 灵感来源于 Process Explorer 的句柄管理功能
