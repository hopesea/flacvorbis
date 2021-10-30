package flacvorbis

import (
	"bytes"
	"strings"

	flac "github.com/go-flac/go-flac"
)

type MetaDataBlockVorbisComment struct {
	Vendor   string
	Comments map[string][]string
}

// New creates a new MetaDataBlockVorbisComment
// vendor is set to flacvorbis <version> by default
func New() *MetaDataBlockVorbisComment {
	return &MetaDataBlockVorbisComment{
		"flacvorbis " + APP_VERSION,
		make(map[string][]string),
	}
}

// Get get all comments with field name specified by the key parameter
// If there is no match, error would still be nil
func (c *MetaDataBlockVorbisComment) Get(key string) ([]string, error) {
	if value, exists := c.Comments[key]; exists {
		return value, nil
	}
	return []string{}, nil
}

// Add adds a key-val pair to the comments
func (c *MetaDataBlockVorbisComment) Add(key string, val string) error {
	for _, char := range key {
		if char < 0x20 || char > 0x7d || char == '=' {
			return ErrorInvalidFieldName
		}
	}
	if xval, exists := c.Comments[key]; exists {
		c.Comments[key] = append(xval, val)
	} else {
		c.Comments[key] = []string{val}
	}
	return nil
}

// Marshal marshals this block back into a flac.MetaDataBlock
func (c MetaDataBlockVorbisComment) Marshal() flac.MetaDataBlock {
	data := bytes.NewBuffer([]byte{})
	packStr(data, c.Vendor)
	data.Write(encodeUint32(uint32(len(c.Comments))))
	for key, values := range c.Comments {
		for _, value := range values {
			packStr(data, key+"="+value)
		}
	}
	return flac.MetaDataBlock{
		Type: flac.VorbisComment,
		Data: data.Bytes(),
	}
}

// ParseFromMetaDataBlock parses an existing picture MetaDataBlock
func ParseFromMetaDataBlock(meta flac.MetaDataBlock) (*MetaDataBlockVorbisComment, error) {
	if meta.Type != flac.VorbisComment {
		return nil, ErrorNotVorbisComment
	}

	reader := bytes.NewReader(meta.Data)
	res := new(MetaDataBlockVorbisComment)

	vendorlen, err := readUint32(reader)
	if err != nil {
		return nil, err
	}
	vendorbytes := make([]byte, vendorlen)
	nn, err := reader.Read(vendorbytes)
	if err != nil {
		return nil, err
	}
	if nn != int(vendorlen) {
		return nil, ErrorUnexpEof
	}
	res.Vendor = string(vendorbytes)

	cmtcount, err := readUint32(reader)
	if err != nil {
		return nil, err
	}
	res.Comments = make(map[string][]string, cmtcount)
	var i uint32
	for i = 0; i < cmtcount; i++ {
		cmtlen, err := readUint32(reader)
		if err != nil {
			return nil, err
		}
		cmtbytes := make([]byte, cmtlen)
		nn, err := reader.Read(cmtbytes)
		if err != nil {
			return nil, err
		}
		if nn != int(cmtlen) {
			return nil, ErrorUnexpEof
		}
		p := strings.SplitN(string(cmtbytes), "=", 2)
		if len(p) != 2 {
			return nil, ErrorMalformedComment
		}

		key, val := p[0], p[1]
		res.Add(key, val)
	}
	return res, nil
}
