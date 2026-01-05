package plugins

import (
	"net/http"
	"sync/atomic"

	"k8s.io/klog/v2"
)

type AtomicHandler struct {
	v atomic.Value
}

func NewAtomicHandler(h http.Handler) *AtomicHandler {
	ah := &AtomicHandler{}
	ah.v.Store(h)
	return ah
}

func (h *AtomicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v := h.v.Load()
	if v == nil {
		klog.Warning("AtomicHandler: no handler stored, returning 503")
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	handler, ok := v.(http.Handler)
	if !ok {
		klog.Warning("AtomicHandler: stored value is not an http.Handler, returning 503")
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	handler.ServeHTTP(w, r)
}

func (h *AtomicHandler) Store(handler http.Handler) {
	h.v.Store(handler)
}
