package cell_base

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/network"
	"github.com/playnb/mustang/network/msg_processor"
	"github.com/playnb/mustang/util"
	"github.com/playnb/protocol"
	"github.com/playnb/protocol/msg"
	"sync"
	"time"
)

type MsgProcessor struct {
	proc network.IMsgProcessor
}

var _mp *MsgProcessor
var _mpOne sync.Once

func getMsgProcessor() *MsgProcessor {
	_mpOne.Do(func() {
		_mp = &MsgProcessor{}
		_mp.proc = msg_processor.NewPbProcessor()
		protocol.KnowAllMsg(_mp)
	})
	return _mp
}

func (mp *MsgProcessor) KnowMsg(sample interface{}) {
	mp.proc.Register(sample, nil)
}

func PackMsg(msg network.MarshalToProto) *util.BuffData {
	data, err := getMsgProcessor().proc.PackMsg(msg)
	if err != nil {
		return nil
	}
	return data
}

func PackForwardToUser(uid uint64, uids []uint64, data network.MarshalToProto, processor network.IMsgProcessor) *msg.MSG_ForwardToUser {
	bufData, err := processor.PackMsg(data)
	if err != nil {
		log.Debug(err.Error())
		return nil
	}

	forward := &msg.MSG_ForwardToUser{}
	forward.UserID = uid //proto.Uint64(uid)
	forward.UserIDs = uids
	forward.CmdData = bufData.Data()
	forward.CreateTime = time.Now().UnixNano() //proto.Int64(time.Now().UnixNano())
	return forward
}

func UnpackForwardToUser(cmd *msg.MSG_ForwardToUser) *util.BuffData {
	bd := util.MakeBuffDataBySlice(cmd.GetCmdData(), 0)
	bd.ChangeIndex(network.BufHeadOffset)
	return bd
}
