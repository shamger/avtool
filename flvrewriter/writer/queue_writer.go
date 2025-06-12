package writer

import (
	"context"
	flv_helper "flvrewriter/flv/helper"
	"flvrewriter/utils"
	"fmt"
	"log"
)

type chanMessage struct {
	gopSize         int
	newKeyTimestamp uint32
}

type queueWriter struct {
	fileWriter *fileWriter
	queue      *utils.Queue
	curTag     *flv_helper.FlvTag
	gopCh      chan *chanMessage
}

func NewQueueWriter(ctx context.Context, outFileName string) Writer {
	queueWriter := &queueWriter{
		fileWriter: NewFileWriter(outFileName).(*fileWriter),
		queue:      utils.NewQueue(),
		curTag:     &flv_helper.FlvTag{},
		gopCh:      make(chan *chanMessage),
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
			case msg := <-w.gopCh:
				gopNum := msg.gopSize
				notInscreasingCnt := 0
				notInscreasingLogSample := ""
				log.Printf("start to flush one gop: %d", gopNum)
				for i := 0; i < gopNum; i++ {
					curTimestamp := w.queue.Peek().(*flv_helper.FlvTag).GetTimestamp()
					if curTimestamp >= msg.newKeyTimestamp {
						notInscreasingCnt++
						notInscreasingLogSample = fmt.Sprintf("timestamp is not increasing, %d > %d",
							curTimestamp, msg.newKeyTimestamp)
						//continue // 暂时不调整写入顺序
					}
					flvTag, ok := w.queue.Dequeue().(*flv_helper.FlvTag)
					if !ok {
						continue
					}
					w.WriteData(flvTag.Header.Bytes())
					w.WriteData(flvTag.Data.Bytes())
					//log.Printf("write tag header: % x", flvTag.Header.Bytes())
				}
				log.Printf("flush one gop done, %s, repeat %d times", notInscreasingLogSample, notInscreasingCnt)
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
		w.gopCh <- &chanMessage{
			gopSize:         w.queue.Size() - 1,
			newKeyTimestamp: w.curTag.GetTimestamp(),
		}
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
