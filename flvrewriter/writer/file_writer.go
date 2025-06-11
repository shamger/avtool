package writer

import (
	"log"
	"os"
)

type fileWriter struct {
	outFileName string
	outFile     *os.File

	curFileSize int64
}

func NewFileWriter(outFileName string) Writer {
	file, err := os.OpenFile(outFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
		return nil
	}
	return &fileWriter{
		outFileName: outFileName,
		outFile:     file,
	}
}

func (w *fileWriter) WriteData(b []byte) (int, error) {
	return w.outFile.Write(b)
}

func (w *fileWriter) WriteTagHeader(b []byte) (int, error) {
	return w.outFile.Write(b)
}

func (w *fileWriter) AppendTagData(b []byte) (int, error) {
	return w.outFile.Write(b)
}

func (w *fileWriter) FinishTagData() {
	fileInfo, err := w.outFile.Stat()
	if err != nil {
		log.Fatalf("stat file failed: %v", err)
		return
	}
	w.curFileSize = fileInfo.Size()
}

func (w *fileWriter) AlignEntireTag() {
	if err := w.outFile.Truncate(w.curFileSize); err != nil {
		log.Fatalf("Truncate failed: %v", err)
		return
	}
	log.Printf("Erase broken tag success")
}

func (w *fileWriter) Seek(offset int64, whence int) (int64, error) {
	return w.outFile.Seek(offset, whence)
}

func (w *fileWriter) Close() error {
	return w.outFile.Close()
}

func (w *fileWriter) GetName() string {
	return w.outFile.Name()
}
