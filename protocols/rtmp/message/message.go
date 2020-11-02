package message

import (
	"bufio"
	"math"

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
	m.MessageStreamID = previousMsg.MessageStreamID
	m.MessageLength = previousMsg.MessageLength
}

func (m *Message) isFinished() bool {
	return (m.currentDataCellPos >= uint32(len(m.MessageData)))
}

func (m *Message) getChunkSize() uint32 {
	chunkSize := m.chunkSize
	if m.currentDataCellPos+chunkSize > uint32(len(m.MessageData)) {
		chunkSize = uint32(len(m.MessageData)) - m.currentDataCellPos
	}
	return chunkSize
}

func (m *Message) firstChunk() error {
	firstChunk := chunk.NewChunk()

	logrus.Debugf("[Debug] Reading FirstChunk of Message")
	if err := firstChunk.Read(m.ReaderWriter.Reader); err != nil {
		return err
	}

	if firstChunk.BasicHeader.Fmt == 1 {
		m.MessageLength = firstChunk.MessageHeader.MessageLength
		m.MessageTypeID = firstChunk.MessageHeader.MessageTypeID
		m.MessageData = make([]byte, m.MessageLength)
	} else if firstChunk.BasicHeader.Fmt == 2 {
		m.MessageTypeID = firstChunk.MessageHeader.MessageTypeID
		m.MessageData = make([]byte, m.MessageLength)
	} else {
		m.MessageStreamID = firstChunk.MessageHeader.MessageStreamID
		m.MessageLength = firstChunk.MessageHeader.MessageLength
		m.MessageTypeID = firstChunk.MessageHeader.MessageTypeID
		m.MessageData = make([]byte, firstChunk.MessageHeader.MessageLength)
	}

	chunkSize := m.getChunkSize()

	logrus.Debugf("[Debug] send buffer to chunk to read with size %v and start point %v and endpoint %v",
		chunkSize, m.currentDataCellPos, m.currentDataCellPos+chunkSize)

	if err := firstChunk.ReadPayload(m.ReaderWriter.Reader,
		m.MessageData[m.currentDataCellPos:m.currentDataCellPos+chunkSize]); err != nil {
		return err
	}
	m.currentDataCellPos += chunkSize

	m.Chunks[m.index] = firstChunk
	m.index++
	logrus.Debugf("[Debug] Finished Reading FirstChunk of Message")

	return nil
}

func (m *Message) Read() error {
	if err := m.firstChunk(); err != nil {
		return err
	}

	for {
		if m.isFinished() {
			m.conn.AckReceived += m.MessageLength
			logrus.Debugf("[Debug] Message Data is %v", m.MessageData)
			logrus.Debug("[Debug] Finished Reading Message")
			break
		}

		curChunk := chunk.NewChunk()
		if err := curChunk.Read(m.ReaderWriter.Reader); err != nil {
			return err
		}

		chunkSize := m.getChunkSize()

		logrus.Debugf("[Debug] Send Buffer with size %v and start index %v and end index %v to Chunk",
			chunkSize, m.currentDataCellPos, m.currentDataCellPos+chunkSize)

		if err := curChunk.ReadPayload(m.ReaderWriter.Reader,
			m.MessageData[m.currentDataCellPos:m.currentDataCellPos+chunkSize]); err != nil {
			return err
		}
		m.currentDataCellPos += chunkSize

		m.Chunks[m.index] = curChunk
		m.index++
	}

	return nil
}

//WriteWithProvidedChunkList is function that write message with provided chunk list in message struct
func (m *Message) WriteWithProvidedChunkList() error {
	for i := range m.Chunks {
		logrus.Debug(m.chunkSize)

		curDataCellPos := i * m.chunkSize
		chunkSize := m.chunkSize
		if curDataCellPos+m.chunkSize > uint32(len(m.MessageData)) {
			chunkSize = uint32(len(m.MessageData)) - curDataCellPos
		}

		if err := m.Chunks[i].Write(m.ReaderWriter.Writer,
			m.MessageData[curDataCellPos:curDataCellPos+chunkSize]); err != nil {
			return err
		}
	}
	return nil
}

//CreateChunksBasedOnFirstChunkThenWrite is function that create ChunkBased on First Chunk Then write Message
func (m *Message) CreateChunksBasedOnFirstChunkThenWrite(Chunk *chunk.Chunk) error {

	numberOfChunks := uint32(math.Ceil(float64(len(m.MessageData)) / float64(m.chunkSize)))

	logrus.Debugf("[Debug] Number of chunk is %v", numberOfChunks)
	chunks := []*chunk.Chunk{}
	chunks = append(chunks, Chunk)
	for i := uint32(1); i < numberOfChunks; i++ {
		chunk := chunk.NewChunk()
		chunk.BasicHeader.Fmt = 3
		chunk.BasicHeader.CSID = Chunk.BasicHeader.CSID
		chunk.MessageHeader.MessageTypeID = Chunk.MessageHeader.MessageTypeID
		chunk.MessageHeader.MessageLength = Chunk.MessageHeader.MessageLength
		chunk.MessageHeader.MessageStreamID = Chunk.MessageHeader.MessageStreamID
		chunk.MessageHeader.TtimeStamp = 0
		chunks = append(chunks, chunk)
	}

	for i := uint32(0); i < numberOfChunks; i++ {
		m.Chunks[i] = chunks[i]
	}

	return m.WriteWithProvidedChunkList()
}
