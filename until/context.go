package until

import (
	"time"

	"golang.org/x/net/context"
)

//TimeOutContext 返回一个带有超时的context
func TimeOutContext(timeout time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return ctx
}

//StandardTimeOutContext 标准超时context
func StandardTimeOutContext() context.Context {
	return TimeOutContext(time.Second * 5)
}
