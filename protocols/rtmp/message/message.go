package message

import (
	"bufio"

	chunk "../ref-chunk"
)

//Message is sequence of chunks
type Message struct {
	MessageStreamID    uint32
	MessageLength      uint32
	MessageTypeID      uint32
	Chunks             map[uint32]*chunk.Chunk
	ReaderWriter       *bufio.ReadWriter
	chunkSize          uint32
	MessageData        []byte
	currentDataCellPos uint32
	index              uint32
}

//NewMessage return mesage with given readerWriter and chunkSize
func NewMessage(readerWriter *bufio.ReadWriter, chunkSize uint32) *Message {
	return &Message{
		ReaderWriter:       readerWriter,
		chunkSize:          chunkSize,
		index:              0,
		currentDataCellPos: 0,
	}
}

func (m *Message) initMessage() {
	firstChunk := chunk.NewChunk()

	chunkSize := m.chunkSize
	if m.currentDataCellPos+chunkSize > uint32(len(m.MessageData)) {
		chunkSize = uint32(len(m.MessageData)) - m.currentDataCellPos
	}
	firstChunk.Read(m.ReaderWriter.Reader, m.MessageData[m.currentDataCellPos:m.currentDataCellPos+chunkSize])
	m.currentDataCellPos += chunkSize

	m.MessageStreamID = firstChunk.MessageHeader.MessageStreamID
	m.MessageLength = firstChunk.MessageHeader.MessageLength
	m.MessageTypeID = firstChunk.MessageHeader.MessageTypeID
	m.MessageData = make([]byte, m.MessageLength)

	m.Chunks[m.index] = firstChunk
	m.index++
}

func (m *Message) isFinished() bool {
	return (m.currentDataCellPos >= uint32(len(m.MessageData)))
}

func (m *Message) Read() {
	m.initMessage()
	for {
		if m.isFinished() {
			break
		}

		curChunk := chunk.NewChunk()
		chunkSize := m.currentDataCellPos + m.chunkSize
		if chunkSize > uint32(len(m.MessageData)) {
			chunkSize = uint32(len(m.MessageData)) - m.currentDataCellPos
		}
		curChunk.Read(m.ReaderWriter.Reader, m.MessageData[m.currentDataCellPos:m.currentDataCellPos+chunkSize])
		m.currentDataCellPos += chunkSize

		m.Chunks[m.index] = curChunk
		m.index++
	}
}

func (m *Message) Write() {
	for i := range m.Chunks {

		curDataCellPos := i * m.chunkSize
		chunkSize := m.chunkSize
		if curDataCellPos+m.chunkSize > uint32(len(m.MessageData)) {
			chunkSize = uint32(len(m.MessageData)) - curDataCellPos
		}

		m.Chunks[i].Write(m.ReaderWriter.Writer, m.MessageData[curDataCellPos:curDataCellPos+chunkSize])
	}
}
