# D2R MultiAblo - Diablo 2 Resurrected 多实例工具

[English](../../README.md) | [繁體中文](README.zh-TW.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md)

适用于 Windows 的《暗黑破坏神 II：重制版》(D2R) 多实例辅助工具，使用 Go 语言开发。

_注意：此文件将在主分支的 README.md 更新后由 ai-translate-action 自动翻译。_

## 下载

您可以从 [Releases](https://github.com/chenwei791129/multiablo/releases) 页面下载最新版本的 Multiablo，或使用下方的直接下载链接：

[🚀 下载最新版 multiablo.exe](https://github.com/chenwei791129/multiablo/releases/latest/download/multiablo.exe)

## 概述

Multiablo 让您能够同时运行多个《暗黑破坏神 II：重制版》实例，通过持续监控并关闭 D2R 用来防止多开的单实例事件句柄。只需在后台运行此工具，您就可以从 Battle.net 启动器启动多个 D2R 实例，无需任何额外步骤。

## 工作原理

D2R 通过创建名为 `DiabloII Check For Other Instances` 的 Windows 事件句柄来防止多开。当 D2R 启动时，它会检查此句柄是否存在 - 如果存在，游戏就会拒绝启动。

Multiablo 的工作方式：
1. 持续监控正在运行的 D2R.exe 进程
2. 自动检测并关闭 `DiabloII Check For Other Instances` 事件句柄
3. 让您可以随时从 Battle.net 启动器启动多个 D2R 实例
4. 监控 `Agent.exe` 进程，仅在其运行超过 7 秒后才终止，最大化 Battle.net 启动器执行"开始游戏"的可用时间

## 使用说明

### 基本用法

1. **运行 multiablo.exe**
2. **从 Battle.net 启动器启动 D2R**
3. **从其他 Battle.net 启动器启动额外的 D2R 实例**！

应用程序启动后会自动开始监控。您可以在 GUI 界面中查看检测到的进程状态和句柄操作结果。

### 杀毒软件误报

某些杀毒软件可能会将 Multiablo 标记为威胁，因为它会操作进程句柄。这是此类工具的预期行为。您可能需要添加例外。

## 免责声明

此工具仅供教育和个人使用。使用风险自负。作者不对以下情况负责：
- 任何违反《暗黑破坏神 II：重制版》服务条款的行为
- 账号停权或封禁
- 游戏崩溃或数据丢失
- 因使用本软件而产生的任何其他问题

**注意**：运行多个实例可能违反游戏的服务条款。使用前请查看暴雪的政策。

## 授权

MIT 授权 - 详见 LICENSE 文件

## 致谢

- 灵感来自 Process Explorer 的句柄管理功能
