package main

import (
	"github.com/playnb/mustang/rpc/example/testrpc"
	"fmt"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
)

var addr = "127.0.0.1:9527"
var wg sync.WaitGroup

type MyEchoService struct {
}

func (s *MyEchoService) Echo(in *testrpc.RequestEcho, out *testrpc.ReplyEcho) error {
	out.Text = proto.String("Echo: ====>" + in.GetText())
	return nil
}

func main() {
	go ListenAndServeEchoService(addr, &MyEchoService{})

	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(id int) {
			client, _ := DialEchoService(addr)
			text := proto.String(fmt.Sprintf("Hello %d", id))
			for i := 0; i < 100; i++ {
				out, _ := client.Echo(&testrpc.RequestEcho{
					Text: text,
				})
				fmt.Println(out.GetText())
				time.Sleep(time.Millisecond)
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}
