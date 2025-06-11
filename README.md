音视频技术学习项目
- flvrewriter
1. 默认: 按tag读取flv文件并重新写入flv文件，meta data添加duration（如果需要）
2. -cp: 支持按tag需要只复制文件前面的部分tag
- flvdumper
1. 默认: 拉取http-flv直播流解析并直接保存成flv文件，会重置时间戳
2. -raw: 直接拉取http-flv直播流并保存二进制文件
3. -queue: 通过tag队列每次写入完整tag
- liverecorder
1. 直播间录制，支持：抖音
