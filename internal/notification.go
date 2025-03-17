package internal

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/containrrr/shoutrrr"
	"github.com/rs/zerolog/log"
)

func SendNotification(surl, message string, proxy string) error {
	// 保存原始的 Transport
	originalTransport := http.DefaultClient.Transport

	if proxy != "" {
		log.Info().Str("Proxy", proxy).Msg("使用代理")
		proxyurl, err := url.Parse(proxy)
		if err != nil {
			log.Fatal().Msgf("代理地址解析失败: %s", err)
		}

		http.DefaultClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyurl),
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}

	err := shoutrrr.Send(surl, message)

	// 还原原始的 Transport
	http.DefaultClient.Transport = originalTransport

	return err
}
