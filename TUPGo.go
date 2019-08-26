package TUPGo

import (
	"encoding/binary"
	"errors"
	"reflect"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"

	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

// PacketBuffer store data
type PacketBuffer struct {
	IVersion int16
	Data     map[string]map[string][]byte // for version 2
	NewData  map[string][]byte            // for version 3
	iRet     int32                        // function ret in response
}

// GetRawData get raw data
func (pb *PacketBuffer) GetRawData() (interface{}, error) {
	if pb.IVersion == 2 {
		return pb.Data, nil
	} else if pb.IVersion == 3 {
		return pb.NewData, nil
	}
	return nil, errors.New("NOT SUPPORT")
}

func (pb *PacketBuffer) init() {
	pb.IVersion = 2
	pb.NewData = make(map[string][]byte)
	pb.Data = make(map[string]map[string][]byte)
}

func (pb *PacketBuffer) get(key string, value interface{}) error {
	if pb.IVersion != 2 && pb.IVersion != 3 {
		return errors.New("NOT SUPPORTED VERSION")
	}
	if key == "" {
		value = pb.iRet
		return nil
	}

	if pb.IVersion == 2 {
		className := reflect.TypeOf(value).Elem().String()
		if val, ok := pb.Data[key]; ok {
			if val1, ok1 := val[className]; ok1 {
				is := codec.NewReader(val1)
				receiver := reflect.ValueOf(value)
				arg := reflect.ValueOf(is)
				f, have := reflect.TypeOf(value).MethodByName("ReadFrom")
				if have {
					f.Func.Call([]reflect.Value{receiver, arg})
				} else {
					return errors.New("NO ReadFrom SUPPORT")
				}
			}
		}
	} else if pb.IVersion == 3 {
		if val, ok := pb.NewData[key]; ok {
			is := codec.NewReader(val)
			receiver := reflect.ValueOf(value)
			arg := reflect.ValueOf(is)
			f, have := reflect.TypeOf(value).MethodByName("ReadFrom")
			if have {
				f.Func.Call([]reflect.Value{receiver, arg})
			} else {
				return errors.New("NO ReadFrom SUPPORT")
			}
		}
	}
	return nil
}

func (pb *PacketBuffer) put(key string, value interface{}) error {
	if pb.IVersion != 2 && pb.IVersion != 3 {
		return errors.New("NOT SUPPORTED VERSION")
	}
	os := codec.NewBuffer()
	receiver := reflect.ValueOf(value)
	arg := reflect.ValueOf(os)
	f, have := reflect.TypeOf(value).MethodByName("WriteTo")
	if !have {
		return errors.New("NO WriteTo SUPPORT")
	}

	f.Func.Call([]reflect.Value{receiver, arg})
	bs := os.ToBytes()

	if pb.IVersion == 2 {
		className := reflect.TypeOf(value).Elem().String()
		if pb.Data[key] == nil {
			pb.Data[key] = make(map[string][]byte)
		}
		if pb.Data[key][className] == nil {
			pb.Data[key][className] = make([]byte, len(bs))
		}
		pb.Data[key][className] = bs
	} else if pb.IVersion == 3 {
		pb.NewData[key] = make([]byte, len(bs))
		pb.NewData[key] = bs
	}
	return nil
}

func (pb *PacketBuffer) writeTo(os *codec.Buffer) error {
	var err error
	os.Reset()
	if pb.IVersion == 2 {
		err = os.WriteHead(codec.MAP, 0)
		if err != nil {
			return err
		}
		err = os.Write_int32(int32(len(pb.Data)), 0)
		if err != nil {
			return err
		}
		for k, v := range pb.Data {
			err = os.Write_string(k, 0)
			if err != nil {
				return err
			}
			err = os.WriteHead(codec.MAP, 1)
			if err != nil {
				return err
			}
			err = os.Write_int32(int32(len(v)), 0)
			if err != nil {
				return err
			}
			for className, value := range v {
				err = os.Write_string(className, 0)
				if err != nil {
					return err
				}
				err = os.WriteHead(codec.SIMPLE_LIST, 1)
				if err != nil {
					return err
				}
				err = os.WriteHead(codec.BYTE, 0)
				if err != nil {
					return err
				}
				err = os.Write_int32(int32(len(value)+2), 0)
				if err != nil {
					return err
				}
				err = os.WriteHead(codec.STRUCT_BEGIN, 0)
				if err != nil {
					return err
				}
				err = os.Write_slice_int8(tools.ByteToInt8(value))
				if err != nil {
					return err
				}
				err = os.WriteHead(codec.STRUCT_END, 0)
				if err != nil {
					return err
				}
			}
		}
	} else if pb.IVersion == 3 {
		err = os.WriteHead(codec.MAP, 0)
		if err != nil {
			return err
		}
		err = os.Write_int32(int32(len(pb.NewData)), 0)
		if err != nil {
			return err
		}
		for k, v := range pb.NewData {
			err = os.Write_string(k, 0)
			if err != nil {
				return err
			}
			err = os.WriteHead(codec.SIMPLE_LIST, 1)
			if err != nil {
				return err
			}

			err = os.WriteHead(codec.BYTE, 0)
			if err != nil {
				return err
			}
			err = os.Write_int32(int32(len(v)+2), 0) // struct begin + struct end
			if err != nil {
				return err
			}
			err = os.WriteHead(codec.STRUCT_BEGIN, 0)
			if err != nil {
				return err
			}
			err = os.Write_slice_int8(tools.ByteToInt8(v))
			if err != nil {
				return err
			}
			err = os.WriteHead(codec.STRUCT_END, 0)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (pb *PacketBuffer) readFrom(is *codec.Reader) error {
	var err error
	var length int32

	if pb.IVersion == 2 {
		err, _ = is.SkipTo(codec.MAP, 0, false)
		if err != nil {
			return err
		}
		err = is.Read_int32(&length, 0, false)
		if err != nil {
			return err
		}

		for i1, e1 := int32(0), length; i1 < e1; i1++ {
			var k1 string

			err = is.Read_string(&k1, 0, false)
			if err != nil {
				return err
			}

			if k1 == "" { // for func ret
				err, _ = is.SkipTo(codec.MAP, 1, false)
				if err != nil {
					return err
				}
				var mapLength int32
				err = is.Read_int32(&mapLength, 0, false)
				if err != nil {
					return err
				}

				for i2, e2 := int32(0), mapLength; i2 < e2; i2++ {
					var k2 string // int32
					err = is.Read_string(&k2, 0, false)
					if err != nil {
						return err
					}

					var retLength int32
					err, _ = is.SkipTo(codec.SIMPLE_LIST, 1, false)
					if err != nil {
						return err
					}
					err, _ = is.SkipTo(codec.BYTE, 0, false)
					if err != nil {
						return err
					}
					err := is.Read_int32(&retLength, 0, false)
					if err != nil {
						return err
					}
					err = is.Read_int32(&pb.iRet, 0, false)
					if err != nil {
						return err
					}
				}
			} else {
				err, _ = is.SkipTo(codec.MAP, 1, false)
				if err != nil {
					return err
				}
				var mapLength int32
				err = is.Read_int32(&mapLength, 0, false)
				if err != nil {
					return err
				}

				for i2, e2 := int32(0), mapLength; i2 < e2; i2++ {
					var k2 string
					var v []int8
					var v1 []byte

					err = is.Read_string(&k2, 0, false)
					if err != nil {
						return err
					}

					err, _ = is.SkipTo(codec.SIMPLE_LIST, 1, false)
					if err != nil {
						return err
					}

					err, _ = is.SkipTo(codec.BYTE, 0, false)
					if err != nil {
						return err
					}

					var byteLength int32

					err = is.Read_int32(&byteLength, 0, false)
					if err != nil {
						return err
					}

					err, _ = is.SkipTo(codec.STRUCT_BEGIN, 0, false)
					if err != nil {
						return err
					}

					err = is.Read_slice_int8(&v, byteLength, false)
					if err != nil {
						return err
					}

					v1 = tools.Int8ToByte(v)

					if pb.Data[k1] == nil {
						pb.Data[k1] = make(map[string][]byte)
					}
					if pb.Data[k1][k2] == nil {
						pb.Data[k1][k2] = make([]byte, byteLength)
					}
					pb.Data[k1][k2] = v1
				}
			}
		}

	} else if pb.IVersion == 3 {
		err, _ = is.SkipTo(codec.MAP, 0, false)
		if err != nil {
			return err
		}
		err = is.Read_int32(&length, 0, false)
		if err != nil {
			return err
		}

		for i1, e1 := int32(0), length; i1 < e1; i1++ {
			var k1 string
			var v []int8
			var v1 []byte

			err = is.Read_string(&k1, 0, false)
			if err != nil {
				return err
			}

			if k1 == "" { // for func ret
				var retLength int32
				err, _ = is.SkipTo(codec.SIMPLE_LIST, 1, false)
				if err != nil {
					return err
				}
				err, _ = is.SkipTo(codec.BYTE, 0, false)
				if err != nil {
					return err
				}
				err := is.Read_int32(&retLength, 0, false)
				if err != nil {
					return err
				}

				err = is.Read_int32(&pb.iRet, 0, false)
				if err != nil {
					return err
				}
			} else {
				err, _ = is.SkipTo(codec.SIMPLE_LIST, 1, false)
				if err != nil {
					return err
				}

				err, _ = is.SkipTo(codec.BYTE, 0, false)
				if err != nil {
					return err
				}

				var byteLength int32

				err = is.Read_int32(&byteLength, 0, false)
				if err != nil {
					return err
				}

				err, _ = is.SkipTo(codec.STRUCT_BEGIN, 0, false)
				if err != nil {
					return err
				}

				err = is.Read_slice_int8(&v, byteLength, false)
				if err != nil {
					return err
				}

				v1 = tools.Int8ToByte(v)

				if pb.NewData[k1] == nil {
					pb.NewData[k1] = make([]byte, byteLength)
				}
				pb.NewData[k1] = v1
			}

		}
	}
	return nil
}

// TarsUniPacket - Go implement
type TarsUniPacket struct {
	IVersion     int16
	CPacketType  int8
	IMessageType int32
	IRequestId   int32
	SServantName string
	SFuncName    string
	Buffer       PacketBuffer
	ITimeout     int32
	Context      map[string]string // req only
	Status       map[string]string // rsp only
}

// Init for TarsUniPacket
func (tup *TarsUniPacket) Init() {
	tup.IVersion = 2
	tup.Buffer.init()
}

// SetVersion for TarsUniPacket
func (tup *TarsUniPacket) SetVersion(version int16) {
	tup.IVersion = version
	tup.Buffer.IVersion = version
}

// SetPacketType for TarsUniPacket
func (tup *TarsUniPacket) SetPacketType(packageType int8) {
	tup.CPacketType = packageType
}

// SetMessageType for TarsUniPacket
func (tup *TarsUniPacket) SetMessageType(messageType int32) {
	tup.IMessageType = messageType
}

// SetRequestId for TarsUniPacket
func (tup *TarsUniPacket) SetRequestId(requestId int32) {
	tup.IRequestId = requestId
}

// SetServantName for TarsUniPacket
func (tup *TarsUniPacket) SetServantName(name string) {
	tup.SServantName = name
}

// SetFuncName for TarsUniPacket
func (tup *TarsUniPacket) SetFuncName(name string) {
	tup.SFuncName = name
}

// SetTimeout for TarsUniPacket
func (tup *TarsUniPacket) SetTimeout(timeout int32) {
	tup.ITimeout = timeout
}

// SetContext for TarsUniPacket
func (tup *TarsUniPacket) SetContext(context map[string]string) {
	tup.Context = context
}

// SetStatus for TarsUniPacket
func (tup *TarsUniPacket) SetStatus(status map[string]string) {
	tup.Status = status
}

// Get from TarsUniPacket
func (tup *TarsUniPacket) Get(key string, value interface{}) error {
	return tup.Buffer.get(key, value)
}

// Put into TarsUniPacket
func (tup *TarsUniPacket) Put(key string, value interface{}) error {
	return tup.Buffer.put(key, value)
}

// Decode from []byte
func (tup *TarsUniPacket) Decode(buff []byte) error {
	pack := requestf.RequestPacket{}
	is := codec.NewReader(buff[4:])
	if err := pack.ReadFrom(is); err != nil {
		return err
	}

	tup.IVersion = pack.IVersion
	tup.Buffer.IVersion = pack.IVersion
	tup.CPacketType = pack.CPacketType
	tup.IMessageType = pack.IMessageType
	tup.IRequestId = pack.IRequestId
	tup.SServantName = pack.SServantName
	tup.SFuncName = pack.SFuncName
	tup.ITimeout = pack.ITimeout
	tup.Context = pack.Context
	tup.Status = pack.Status

	is1 := codec.NewReader(tools.Int8ToByte(pack.SBuffer))
	if err := tup.Buffer.readFrom(is1); err != nil {
		return err
	}

	return nil
}

// Encode to []byte
func (tup *TarsUniPacket) Encode() ([]byte, error) {
	if tup.SServantName == "" || tup.SFuncName == "" {
		return nil, errors.New("servantName and funcName is required")
	}
	os := codec.NewBuffer()
	if err := tup.Buffer.writeTo(os); err != nil {
		return nil, err
	}
	pack := requestf.RequestPacket{
		IVersion:     tup.Buffer.IVersion,
		CPacketType:  tup.CPacketType,
		IMessageType: tup.IMessageType,
		IRequestId:   tup.IRequestId,
		SServantName: tup.SServantName,
		SFuncName:    tup.SFuncName,
		SBuffer:      tools.ByteToInt8(os.ToBytes()),
		ITimeout:     tup.ITimeout,
		Context:      tup.Context,
		Status:       tup.Status,
	}

	os1 := codec.NewBuffer()
	if err := pack.WriteTo(os1); err != nil {
		return nil, err
	}

	result := make([]byte, 4)
	result = append(result, os1.ToBytes()...)
	binary.BigEndian.PutUint32(result[:4], uint32(len(result)))

	return result, nil
}
