package proxy

import (
	"net/url"
	"sync/atomic"

	"github.com/opreader/nami/common"
)

type RoundRobinSwitcher struct {
	proxyURLs []*url.URL
	index     uint32
}

func (r *RoundRobinSwitcher) GetProxy() *url.URL {
	index := atomic.AddUint32(&r.index, 1) - 1
	u := r.proxyURLs[index%uint32(len(r.proxyURLs))]
	return u
}

// RoundRobinProxySwitcher creates a proxy switcher function which rotates ProxyURLs on every request.
// The proxy type is determined by the URL scheme. "http", "https" and "socks5" are supported.
// If the scheme is empty, "http" is assumed.
func RoundRobinProxySwitcher(ProxyURLs ...string) (*RoundRobinSwitcher, error) {
	if len(ProxyURLs) < 1 {
		return nil, common.ErrEmptyProxyURL
	}
	urls := make([]*url.URL, len(ProxyURLs))
	for i, u := range ProxyURLs {
		parsedU, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		urls[i] = parsedU
	}
	return &RoundRobinSwitcher{urls, 0}, nil
}
