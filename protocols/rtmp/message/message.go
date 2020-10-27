package message

import (
	"bufio"

	"github.com/sirupsen/logrus"

	"../conn"
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
	conn               *conn.Conn
	prevMsg            *Message
}

//NewMessage return mesage with given readerWriter and chunkSize
func NewMessage(readerWriter *bufio.ReadWriter, chunkSize uint32, conn *conn.Conn) *Message {
	return &Message{
		ReaderWriter:       readerWriter,
		chunkSize:          chunkSize,
		index:              0,
		currentDataCellPos: 0,
		Chunks:             make(map[uint32]*chunk.Chunk),
		conn:               conn,
	}
}

//FetchWithPreviousMsg store prevous msg for get information for msg type 1 and 2
func (m *Message) FetchWithPreviousMsg(previousMsg *Message) {
	logrus.Debug("[Debug] Fetching Message")
	m.prevMsg = previousMsg
}

func (m *Message) initMessage() {
	firstChunk := chunk.NewChunk()

	logrus.Debugf("[Debug] init message")
	chunkSize := m.chunkSize
	firstChunk.Read(m.ReaderWriter.Reader)

	if firstChunk.BasicHeader.Fmt == 1 {
		m.MessageStreamID = m.prevMsg.MessageStreamID
		m.MessageLength = firstChunk.MessageHeader.MessageLength
		m.MessageTypeID = firstChunk.MessageHeader.MessageTypeID
		m.MessageData = make([]byte, m.MessageLength)
	} else if firstChunk.BasicHeader.Fmt == 2 {
		logrus.Debugf("[Debug] size of previous msg is %v", m.prevMsg.MessageLength)
		m.MessageStreamID = m.prevMsg.MessageStreamID
		m.MessageLength = m.prevMsg.MessageLength
		m.MessageTypeID = firstChunk.MessageHeader.MessageTypeID
		m.MessageData = make([]byte, m.MessageLength)
	} else {
		m.MessageStreamID = firstChunk.MessageHeader.MessageStreamID
		m.MessageLength = firstChunk.MessageHeader.MessageLength
		m.MessageTypeID = firstChunk.MessageHeader.MessageTypeID
		m.MessageData = make([]byte, firstChunk.MessageHeader.MessageLength)
	}

	if m.currentDataCellPos+chunkSize > uint32(len(m.MessageData)) {
		chunkSize = uint32(len(m.MessageData)) - m.currentDataCellPos
	}

	logrus.Debugf("[Debug] send buffer to chunk to read with size %v and start point %v and endpoint %v",
		chunkSize, m.currentDataCellPos, m.currentDataCellPos+chunkSize)

	firstChunk.ReadPayload(m.ReaderWriter.Reader, m.MessageData[m.currentDataCellPos:m.currentDataCellPos+chunkSize])
	m.currentDataCellPos += chunkSize

	m.Chunks[m.index] = firstChunk
	m.index++
	logrus.Debugf("[Debug] init message finished")
}

func (m *Message) isFinished() bool {
	return (m.currentDataCellPos >= uint32(len(m.MessageData)))
}

func (m *Message) Read() {
	m.initMessage()
	for {
		if m.isFinished() {
			m.conn.AckReceived += m.MessageLength
			logrus.Debugf("[Debug] Message Data is %v", m.MessageData)
			logrus.Debug("[Debug] finished reading message")
			break
		}

		curChunk := chunk.NewChunk()
		curChunk.Read(m.ReaderWriter.Reader)

		chunkSize := m.currentDataCellPos + m.chunkSize
		if chunkSize > uint32(len(m.MessageData)) {
			chunkSize = uint32(len(m.MessageData)) - m.currentDataCellPos
		}
		logrus.Debugf("[Debug] send buffer to chunk to read with size %v and start point %v and endpoint %v",
			m.chunkSize, m.currentDataCellPos, m.currentDataCellPos+chunkSize)

		curChunk.ReadPayload(m.ReaderWriter.Reader, m.MessageData[m.currentDataCellPos:m.currentDataCellPos+chunkSize])
		m.currentDataCellPos += chunkSize

		m.Chunks[m.index] = curChunk
		m.index++
	}
}

func (m *Message) Write() {
	logrus.Debug(m.Chunks)
	for i := range m.Chunks {
		logrus.Debug(m.chunkSize)

		curDataCellPos := i * m.chunkSize
		chunkSize := m.chunkSize
		if curDataCellPos+m.chunkSize > uint32(len(m.MessageData)) {
			chunkSize = uint32(len(m.MessageData)) - curDataCellPos
		}

		m.Chunks[i].Write(m.ReaderWriter.Writer, m.MessageData[curDataCellPos:curDataCellPos+chunkSize])
	}
}
