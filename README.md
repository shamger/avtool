# AVTool - 音视频技术学习工具集

AVTool 是一个用于学习和研究音视频处理技术的工具集合，包含多个实用的音视频处理工具。

## 项目概述

本项目旨在通过实际代码实现来学习和理解音视频处理的相关技术，包括FLV格式解析、HLS协议、直播流录制等。

## 工具列表

### 1. FLVRewriter - FLV文件重写工具
- **功能描述**: 按tag读取FLV文件并重新写入FLV文件，支持meta data添加duration（如果需要）
- **特殊功能**: 支持`-cp`参数按tag需要只复制文件前面的部分tag

### 2. FLVDumper - FLV流解析工具
- **功能描述**: 拉取HTTP-FLV直播流解析并直接保存成FLV文件，会重置时间戳
- **特殊功能**: 
  - `-raw`: 直接拉取HTTP-FLV直播流并保存二进制文件
  - `-queue`: 通过tag队列每次写入完整tag

### 3. LiveRecorder - 直播录制工具
- **功能描述**: 直播间录制功能
- **支持平台**: 抖音

### 4. HLSDumper - HLS流解析工具
- **功能描述**: HLS流解析和dump功能

## 使用说明

每个工具都有相应的命令行参数和配置选项，具体使用方法请参考各工具的文档或源代码。

## 参考项目

- https://github.com/coreyauger/flv-streamer-2-file
- https://en.wikipedia.org/wiki/Flash_Video

## 免责声明

本项目仅供学习和研究使用，请勿用于商业用途或违反相关法律法规的场景。

**仅学习使用和参考项目**
