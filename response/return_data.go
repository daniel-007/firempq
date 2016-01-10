package encoding

import (
	"io"
	"strconv"
	"strings"

	. "firempq/api"
	. "firempq/encoding"
	. "firempq/utils"
)

type CallFuncType func([]string) IResponse

type DictResponse struct {
	dict   map[string]interface{}
	header string
}

func NewDictResponse(header string, dict map[string]interface{}) *DictResponse {
	return &DictResponse{
		dict:   dict,
		header: header,
	}
}

func (self *DictResponse) GetDict() map[string]interface{} {
	return self.dict
}

func (self *DictResponse) getResponseChunks() []string {
	data := make([]string, 0, 3+9*len(self.dict))
	data = append(data, self.header)
	data = append(data, " %")
	data = append(data, strconv.Itoa(len(self.dict)))
	for k, v := range self.dict {
		data = append(data, " ")
		data = append(data, k)
		data = append(data, " ")
		switch t := v.(type) {
		case string:
			data = append(data, t)
		case int:
			data = append(data, ":")
			data = append(data, strconv.Itoa(t))
		case int64:
			data = append(data, ":")
			data = append(data, strconv.Itoa(int(t)))
		case bool:
			if t {
				data = append(data, "?t")
			} else {
				data = append(data, "?f")
			}
		}
	}
	return data
}

func (self *DictResponse) GetResponse() string {
	return strings.Join(self.getResponseChunks(), "")
}

func (self *DictResponse) WriteResponse(buff io.Writer) error {
	var err error
	for _, s := range self.getResponseChunks() {
		_, err = buff.Write(UnsafeStringToBytes(s))
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *DictResponse) IsError() bool { return false }

type ItemsResponse struct {
	items []IResponseItem
}

func NewItemsResponse(items []IResponseItem) *ItemsResponse {
	return &ItemsResponse{
		items: items,
	}
}

func (self *ItemsResponse) GetItems() []IResponseItem {
	return self.items
}

func (self *ItemsResponse) getResponseChunks() []string {
	data := make([]string, 0, 3+9*len(self.items))
	data = append(data, "+MSGS")
	data = append(data, EncodeArraySize(len(self.items)))
	for _, item := range self.items {
		data = append(data, item.Encode())
	}
	return data
}

func (self *ItemsResponse) GetResponse() string {
	return strings.Join(self.getResponseChunks(), "")
}

func (self *ItemsResponse) WriteResponse(buff io.Writer) error {
	var err error
	for _, s := range self.getResponseChunks() {
		_, err = buff.Write(UnsafeStringToBytes(s))
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *ItemsResponse) IsError() bool {
	return false
}