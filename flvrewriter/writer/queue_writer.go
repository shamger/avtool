package writer

import (
	"context"
	flv_helper "flvrewriter/flv/helper"
	"flvrewriter/utils"
	"log"
)

type queueWriter struct {
	fileWriter *fileWriter
	queue      *utils.Queue
	curTag     *flv_helper.FlvTag
	gopCh      chan int
}

func NewQueueWriter(ctx context.Context, outFileName string) Writer {
	queueWriter := &queueWriter{
		fileWriter: NewFileWriter(outFileName).(*fileWriter),
		queue:      utils.NewQueue(),
		curTag:     &flv_helper.FlvTag{},
		gopCh:      make(chan int),
	}
	queueWriter.startFlushTags(ctx)
	return queueWriter
}

func (w *queueWriter) startFlushTags(ctx context.Context) {
	go func() {
		log.Printf("start to flush tags...")
		for {
			select {
			case <-ctx.Done():
				log.Printf("flushing worker is stopped")
				return
			case size := <-w.gopCh:
				gopNum := size - 1
				log.Printf("start to flush one gop: %d", gopNum)
				for i := 0; i < gopNum; i++ {
					flvTag, ok := w.queue.Dequeue().(*flv_helper.FlvTag)
					if !ok {
						continue
					}
					w.WriteData(flvTag.Header.Bytes())
					w.WriteData(flvTag.Data.Bytes())
					//log.Printf("write tag header: % x", flvTag.Header.Bytes())
				}
				log.Printf("flush one gop done")
			}
		}
	}()
}

func (w *queueWriter) WriteData(b []byte) (int, error) {
	return w.fileWriter.outFile.Write(b)
}

func (w *queueWriter) WriteTagHeader(b []byte) (int, error) {
	w.curTag.Header.Write(b)
	return len(b), nil
}

func (w *queueWriter) AppendTagData(b []byte) (int, error) {
	return w.curTag.Data.Write(b)
}

func (w *queueWriter) FinishTagData() {
	if w.curTag.IsKeyTag() && w.queue.Size() > 0 {
		log.Printf("notify to flush one gop, queue size:%d", w.queue.Size())
		w.gopCh <- w.queue.Size()
	}
	//log.Printf("equeue header: % x", w.curTag.Header.Bytes())
	w.queue.Enqueue(w.curTag)
	w.curTag = &flv_helper.FlvTag{}
}

func (w *queueWriter) AlignEntireTag() {
	log.Printf("no need to align for queue writing")
}

// TODO 进一步封装
func (w *queueWriter) Seek(offset int64, whence int) (int64, error) {
	return w.fileWriter.outFile.Seek(offset, whence)
}

func (w *queueWriter) Close() error {
	return w.fileWriter.Close()
}

func (w *queueWriter) GetName() string {
	return w.fileWriter.GetName()
}
