package mail

import (
	"github.com/playnb/mustang/mail/mail-server/pb"
	"github.com/playnb/mustang/redis"
	"github.com/playnb/mustang/util"
	"cell/conf"
	"context"
	"testing"
	"time"
)

func init() {
	redis.InitRedisPool("127.0.0.1:6379", "")
	conf.GetMe().ConfigDir = `C:\ltp\PPGame\doc`
}

func TestMailNotify(t *testing.T) {
	stream, err := GetClient().GetNewMailNotifications(context.Background())
	if err != nil {
		t.Fail()
	}
	for i := uint64(0); i < 10; i++ {
		stream.Send(&pb.SubscribeRequest{
			IsSubscribe: true,
			To:          util.Uint64ToBytes(200 + i),
		})
	}
	go func() {
		for {
			notify, err := stream.Recv()
			if err != nil {
				t.Fail()
			}
			t.Log("notify")
			for _, v := range notify.ToList {
				t.Logf("Notify:%d\n", util.BytesToUint64(v))
			}
		}
	}()
	for i := uint64(0); i < 20; i++ {
		resp, err := GetClient().Send(context.Background(), &pb.SendRequest{
			Header: &pb.MailHeader{
				From: util.Uint64ToBytes(10),
				To:   util.Uint64ToBytes(200 + i%10),
				Time: &pb.MailHeader_Time{
					Send:   util.NowTimestamp(),
					Expire: util.NowTimestamp() + 86400*15,
				},
				Title: []byte("Try"),
			},
			Body: &pb.MailBody{
				Body: []byte("10===>20"),
			},
		})
		if err != nil {
			t.Fail()
		}
		//t.Log(resp)
		t.Logf("Send:%v\n", resp.GetMailId())
		time.Sleep(time.Second * 1)
	}
	time.Sleep(time.Second * 5)
	return
}

func TestMailClient(t *testing.T) {
	//t.Log(util.Float32ToByte(3.1415926))
	//t.Log(util.BytesToFloat64(util.Float64ToByte(3.123456789)))
	var mailID []byte

	for i := 0; i > 11; i++ {
		resp, err := GetClient().Send(context.Background(), &pb.SendRequest{
			Header: &pb.MailHeader{
				From: util.Uint64ToBytes(10),
				To:   util.Uint64ToBytes(200),
				Time: &pb.MailHeader_Time{
					Send:   util.NowTimestamp(),
					Expire: util.NowTimestamp() + 86400*15,
				},
				Title: []byte("Try"),
			},
			Body: &pb.MailBody{
				Body: []byte("10===>20"),
			},
		})
		if err != nil {
			t.Fail()
		}
		t.Log(resp)
		t.Logf("Send:%v\n", resp.GetMailId())
		mailID = resp.GetMailId()
	}

	if len(mailID) == 12 {
		resp, err := GetClient().Get(context.Background(), &pb.GetRequest{
			MailIndex: &pb.MailIndex{
				To:     util.Uint64ToBytes(200),
				MailId: mailID,
			},
		})
		if err != nil {
			t.Fail()
		}
		t.Log(resp.GetBody())
	}
	{
		listMail := func(req *pb.ListRequest) (*pb.ListResponse, *pb.ListRequest_TimeAndID) {
			t.Log("listMail-------")
			lastOne := &pb.ListRequest_TimeAndID{}
			resp, err := GetClient().List(context.Background(), req)
			if err != nil || !resp.GetResult().GetOk() {
				t.Log(resp.GetResult())
				t.Fail()
				return nil, nil
			}
			for k, v := range resp.GetMails() {
				lastOne.MailId = v.GetMailId()
				lastOne.Time = v.GetHeader().GetTime().GetSend()
				t.Logf("%d:%v\n", k, v.GetMailId())
			}
			t.Logf("%v\n", lastOne)
			return resp, lastOne
		}

		var resp *pb.ListResponse
		var lastOne *pb.ListRequest_TimeAndID
		for {
			if lastOne != nil {
				lastOne.Time = util.NowTimestamp()
			}
			resp, lastOne = listMail(&pb.ListRequest{
				To:     util.Uint64ToBytes(200),
				Limit:  5,
				End:    lastOne,
				Ascend: false,
			})
			if len(resp.GetMails()) < 5 {
				break
			}
		}
	}
}
