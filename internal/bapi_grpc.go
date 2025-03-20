package internal

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	bilimetadata "github.com/XiaoMiku01/bilibili-grpc-api-go/bilibili/metadata"
	"github.com/XiaoMiku01/bilibili-grpc-api-go/bilibili/metadata/device"
	"github.com/XiaoMiku01/bilibili-grpc-api-go/bilibili/metadata/locale"
	"github.com/XiaoMiku01/bilibili-grpc-api-go/bilibili/metadata/network"
	"github.com/XiaoMiku01/bilibili-grpc-api-go/bilibili/rpc"

	playapi "github.com/XiaoMiku01/bilibili-grpc-api-go/bilibili/app/playurl/v1"
	viewapi "github.com/XiaoMiku01/bilibili-grpc-api-go/bilibili/app/view/v1"
	dmapi "github.com/XiaoMiku01/bilibili-grpc-api-go/bilibili/community/service/dm/v1"
)

type ViewReq = viewapi.ViewReq
type ViewReply = viewapi.ViewReply
type DmSegMobileReq = dmapi.DmSegMobileReq
type DanmakuStruct = dmapi.DanmakuElem

// RetryUnaryInterceptor 创建一个支持条件重试的gRPC一元拦截器
func RetryUnaryInterceptor(maxRetries int, backoffDuration time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var lastErr error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			err := invoker(ctx, method, req, reply, cc, opts...)
			log.Debug().Msgf("Hook Grpc : %s, req: %v, reply: %v, err: %v [%d/%d]", method, req, reply, err, attempt+1, maxRetries)
			if err == nil {
				return nil
			}
			lastErr = formatGRPCError(err)

			if attempt >= maxRetries {
				return err
			}
			waitTime := backoffDuration * time.Duration(attempt+1)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):

			}
		}

		return lastErr
	}
}

// InitGRPC 初始化B站GRPC客户端
func (ba *BApiClient) InitGRPC() error {
	addr := "grpc.biliapi.net:443"
	creds := grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS12,
	}))
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second, // 每10秒发送ping帧保持连接活跃
		Timeout:             10 * time.Second, // 等待pong响应的超时时间
		PermitWithoutStream: true,             // 允许在没有活跃RPC调用的情况下也发送保活包
	}
	options := []grpc.DialOption{
		creds,
		grpc.WithKeepaliveParams(kacp),
		grpc.WithUnaryInterceptor(RetryUnaryInterceptor(3, 1*time.Second)),
	}
	conn, err := grpc.NewClient(addr, options...)
	if err != nil {
		return err
	}

	ba.grpcConn = conn
	ba.dmClient = dmapi.NewDMClient(conn)
	ba.playurlClient = playapi.NewPlayURLClient(conn)
	ba.viewClient = viewapi.NewViewClient(conn)

	return nil
}

// CloseGRPC 关闭GRPC连接
func (ba *BApiClient) CloseGRPC() error {
	if ba.grpcConn != nil {
		return ba.grpcConn.Close()
	}
	return nil
}

// getGRPCMetadata 获取B站GRPC请求元数据
func (ba *BApiClient) getGRPCMetadata() metadata.MD {
	buvid := func() string {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		prefix := "XX"
		const hexChars = "0123456789ABCDEF"
		result := make([]byte, 35)
		for i := range result {
			result[i] = hexChars[r.Intn(len(hexChars))]
		}
		return prefix + string(result)
	}()
	device := &device.Device{
		MobiApp:  "android",
		Device:   "phone",
		Build:    6830300,
		Channel:  "bili",
		Buvid:    buvid,
		Platform: "android",
	}
	devicebin, _ := proto.Marshal(device)
	locale := &locale.Locale{
		Timezone: "Asia/Shanghai",
	}
	localebin, _ := proto.Marshal(locale)
	bilimetadata := &bilimetadata.Metadata{
		AccessKey: ba.accessKey,
		MobiApp:   "android",
		Device:    "phone",
		Build:     6830300,
		Channel:   "bili",
		Buvid:     buvid,
		Platform:  "android",
	}
	bilimetadatabin, _ := proto.Marshal(bilimetadata)
	network := &network.Network{
		Type: network.NetworkType_WIFI,
	}
	networkbin, _ := proto.Marshal(network)
	md := metadata.Pairs(
		"x-bili-device-bin", string(devicebin),
		"x-bili-local-bin", string(localebin),
		"x-bili-metadata-bin", string(bilimetadatabin),
		"x-bili-network-bin", string(networkbin),
		"authorization", "identify_v1 "+ba.accessKey,
	)
	return md
}

// getGRPCContext 获取附带元数据的上下文
func (ba *BApiClient) getGRPCContext() context.Context {
	md := ba.getGRPCMetadata()
	return metadata.NewOutgoingContext(context.Background(), md)
}

// formatGRPCError 格式化GRPC错误
func formatGRPCError(err error) error {
	status, ok := status.FromError(err)
	if !ok {
		return err
	}
	// B站的grpc接口返回的错误码，例如鉴权错误
	if status.Code() == codes.Unknown && len(status.Details()) > 0 {
		rpcStatus, ok := status.Details()[0].(*rpc.Status)
		if ok {
			return fmt.Errorf("B站GRPC错误 %d: %s", rpcStatus.Code, rpcStatus.Message)
		}
	}
	return err
}

// GetDanmaku 获取弹幕
func (ba *BApiClient) GetDanmaku(req *dmapi.DmSegMobileReq) (*dmapi.DmSegMobileReply, error) {
	resp, err := ba.dmClient.DmSegMobile(ba.getGRPCContext(), req)
	if err != nil {
		return nil, formatGRPCError(err)
	}
	return resp, nil
}

// GetView 获取视频信息
func (ba *BApiClient) GetView(req *viewapi.ViewReq) (*viewapi.ViewReply, error) {
	resp, err := ba.viewClient.View(ba.getGRPCContext(), req)
	if err != nil {
		return nil, formatGRPCError(err)
	}
	return resp, nil
}
