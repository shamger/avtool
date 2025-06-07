package writer

// 定义写flv tag接口，仅限重复出现的tag
type Writer interface {
	WriteData(b []byte) (int, error)      // 写入任意二进制序列
	WriteTagHeader(b []byte) (int, error) // 写入完整tag header
	AppendTagData(b []byte) (int, error)  // 追加tag data
	FinishTagData()                       // tag data已完成
	AlignEntireTag()
	Seek(offset int64, whence int) (int64, error)
	Close() error
	GetName() string // 获取最终写入的文件名
}
