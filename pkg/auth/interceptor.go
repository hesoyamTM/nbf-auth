package auth

import (
	"context"
	"encoding/base64"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	UidInterceptor     = "uid"
	NameInterceptor    = "name"
	SurnameInterceptor = "surname"
)

func SettingMetadataInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		uid, ok := ctx.Value(UID).(string)
		if !ok {
			uid = ""
		}
		name, ok := ctx.Value(NAME).(string)
		if !ok {
			name = ""
		}
		surname, ok := ctx.Value(SURNAME).(string)
		if !ok {
			surname = ""
		}

		encodedName := base64.URLEncoding.EncodeToString([]byte(name))
		encodedSurname := base64.URLEncoding.EncodeToString([]byte(surname))

		md := metadata.New(map[string]string{})

		md.Set(UidInterceptor, uid)
		md.Set(NameInterceptor, encodedName)
		md.Set(SurnameInterceptor, encodedSurname)

		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func TakingMetadataInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}

	if uids := md.Get(UidInterceptor); len(uids) > 0 {
		ctx = context.WithValue(ctx, UID, uids[0])
	}
	if names := md.Get(NameInterceptor); len(names) > 0 {
		decodedName, err := base64.URLEncoding.DecodeString(names[0])
		if err != nil {
			return handler(ctx, req)
		}

		ctx = context.WithValue(ctx, NAME, decodedName)
	}
	if surnames := md.Get(SurnameInterceptor); len(surnames) > 0 {
		decodedSurname, err := base64.URLEncoding.DecodeString(surnames[0])
		if err != nil {
			return handler(ctx, req)
		}

		ctx = context.WithValue(ctx, SURNAME, decodedSurname)
	}

	return handler(ctx, req)
}
