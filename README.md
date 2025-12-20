# Multiablo

[English](README.md) | [繁體中文](README.zh-TW.md)

A D2R (Diablo II: Resurrected) multi-instance helper for Windows, written in Go.

## Overview

Multiablo enables you to run multiple instances of Diablo II: Resurrected simultaneously by continuously monitoring and closing the single-instance Event Handle that D2R uses to prevent multi-launching. Simply run this tool in the background, and you can launch multiple D2R instances from the Battle.net launcher without any additional steps.

## How It Works

D2R prevents multiple instances by creating a Windows Event Handle named `DiabloII Check For Other Instances`. When D2R starts, it checks if this handle exists - if it does, the game refuses to launch.

Multiablo works by:
1. Continuously monitoring for running D2R.exe processes
2. Automatically detecting and closing the `DiabloII Check For Other Instances` Event Handle
3. Allowing you to launch multiple D2R instances from Battle.net launcher at any time
4. Continuously monitoring and terminating `Agent.exe` processes that may interfere with multi-instance operation

## Usage

### Basic Usage

1. **Start D2R** from Battle.net launcher
2. **Run multiablo.exe**
3. **Launch additional D2R instances** from other Battle.net launcher!

### Command-Line Options

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

### Example Output

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

### Antivirus False Positive

Some antivirus software may flag Multiablo because it manipulates process handles. This is expected behavior for this type of tool. You may need to add an exception.

## Disclaimer

This tool is for educational and personal use only. Use at your own risk. The authors are not responsible for:
- Any violations of Diablo II: Resurrected Terms of Service
- Account suspensions or bans
- Game crashes or data loss
- Any other issues arising from use of this software

**Note**: Running multiple instances may violate the game's Terms of Service. Check Blizzard's policies before use.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Inspired by Process Explorer's handle management capabilities
