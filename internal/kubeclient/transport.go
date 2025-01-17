package kubeclient

import (
	"net/http"

	netutil "k8s.io/apimachinery/pkg/util/net"
)

const (
	xKargoUserCredentialHeader = "X-Kargo-User-Credential" // #nosec G101
)

var (
	_ netutil.RoundTripperWrapper = &credentialHook{}
	_ netutil.RoundTripperWrapper = &credentialInjector{}
)

type credentialHook struct {
	rt http.RoundTripper
}

func newAuthorizationHeaderHook(rt http.RoundTripper) http.RoundTripper {
	return &credentialHook{
		rt: rt,
	}
}

func (h *credentialHook) RoundTrip(req *http.Request) (*http.Response, error) {
	cred := req.Header.Get("Authorization")
	res, err := h.rt.RoundTrip(req)
	if res != nil {
		res.Header.Set(xKargoUserCredentialHeader, cred)
	}
	return res, err
}

func (h *credentialHook) WrappedRoundTripper() http.RoundTripper {
	return h.rt
}

type credentialInjector struct {
	rt http.RoundTripper
}

func NewCredentialInjector(rt http.RoundTripper) http.RoundTripper {
	return &credentialInjector{
		rt: rt,
	}
}
func (i *credentialInjector) RoundTrip(req *http.Request) (*http.Response, error) {
	cred, ok := GetCredentialFromContext(req.Context())
	if !ok {
		return i.rt.RoundTrip(req)
	}

	newReq := req.Clone(req.Context())
	newReq.Header.Set("Authorization", cred)
	return i.rt.RoundTrip(newReq)
}

func (i *credentialInjector) WrappedRoundTripper() http.RoundTripper {
	return i.rt
}
