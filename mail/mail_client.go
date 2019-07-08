package mail

import (
	"github.com/playnb/mustang/mail/mail-server"
	"github.com/playnb/mustang/mail/mail-server/pb"
	"cell/conf"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"sync"
)

var (
	cm     = &sync.Mutex{}
	client pb.MailClient
)

func GetClient() pb.MailClient {
	cm.Lock()
	defer cm.Unlock()
	if client != nil {
		return client
	}

	creds, err := credentials.NewClientTLSFromFile(conf.GetMe().ConfigDir+"/mail-server-key/server.crt", "gamemail.ztgame.com")
	if err != nil {
		panic(fmt.Errorf("could not load tls cert: %s", err))
	}

	serviceURL := conf.GetMe().MailService

	conn, _ := grpc.Dial(
		serviceURL,
		grpc.WithTransportCredentials(creds),
		// grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(mail_server.TokenAuth{
			Token: "test",
		}),
		/*
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				lile.ContextClientInterceptor(),
				otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer()),
			),
		)*/)

	client = pb.NewMailClient(conn)
	return client
}
