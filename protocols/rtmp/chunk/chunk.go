package chunk

//Chunk is base packet that send over rtmp protocol and have
//2 field and first field is Chunk header second is Chunk data
//Chunk header consist of 3 spereate header:
//1.basic header
//2.message header
//3.extended timestamp
type Chunk struct {

	//basic header
	fmt  uint32
	CSID uint32

	//message header
	timeStamp       uint32
	messageLength   uint32
	messageTypeID   uint32
	messageStreamID uint32
	timeStampDelta  uint32

	haveExtendedTimeStamp bool

	//chunk data
	data []byte
}
