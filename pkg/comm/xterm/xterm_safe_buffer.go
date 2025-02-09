package xterm

import (
	"bytes"
	"sync"
)

// 定义线程安全的缓冲区
type SafeBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (b *SafeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Write(p)
}

func (b *SafeBuffer) Read(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Read(p)
}

func (b *SafeBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Len()
}

func (b *SafeBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Bytes()
}

func (b *SafeBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buffer.Reset()
}
