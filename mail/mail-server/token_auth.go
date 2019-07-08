package mail_server

import "golang.org/x/net/context"

type TokenAuth struct {
	Token string
}

func (t TokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization-token": t.Token,
	}, nil
}

func (TokenAuth) RequireTransportSecurity() bool {
	return true
}
