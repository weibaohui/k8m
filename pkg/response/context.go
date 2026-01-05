package response

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

type H map[string]interface{}

type ContextKey struct{}

var ctxKey = ContextKey{}
var decoder = schema.NewDecoder()

func init() {
	decoder.IgnoreUnknownKeys(true)
}

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

func New(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{Writer: w, Request: r}
}

func FromRequest(r *http.Request) *Context {
	if ctx, ok := r.Context().Value(ctxKey).(*Context); ok {
		return ctx
	}
	return New(nil, r)
}

func (c *Context) JSON(status int, obj interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(status)
	if obj != nil {
		_ = json.NewEncoder(c.Writer).Encode(obj)
	}
}

func (c *Context) String(status int, s string) {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(status)
	_, _ = c.Writer.Write([]byte(s))
}

func (c *Context) Data(status int, contentType string, data []byte) {
	c.Writer.Header().Set("Content-Type", contentType)
	c.Writer.WriteHeader(status)
	_, _ = c.Writer.Write(data)
}

func (c *Context) AbortWithJSON(status int, obj interface{}) {
	c.JSON(status, obj)
}

func (c *Context) AbortWithString(status int, s string) {
	c.String(status, s)
}

func (c *Context) Status(status int) *Context {
	c.Writer.WriteHeader(status)
	return c
}

func (c *Context) ShouldBindJSON(obj interface{}) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, obj)
}

func (c *Context) ShouldBind(obj interface{}) error {
	return c.ShouldBindJSON(obj)
}

func (c *Context) ShouldBindQuery(obj interface{}) error {
	return decoder.Decode(obj, c.Request.URL.Query())
}

func (c *Context) Redirect(code int, url string) {
	http.Redirect(c.Writer, c.Request, url, code)
}

func (c *Context) PostForm(key string) string {
	return c.Request.PostFormValue(key)
}

func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	file, header, err := c.Request.FormFile(key)
	if err != nil {
		return nil, err
	}
	file.Close()
	return header, nil
}

func (c *Context) Flush() {
	if f, ok := c.Writer.(http.Flusher); ok {
		f.Flush()
	}
}

func (c *Context) SSEvent(name string, data interface{}) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	if name != "" {
		fmt.Fprintf(c.Writer, "event: %s\n", name)
	}
	fmt.Fprintf(c.Writer, "data: %v\n\n", data)
	c.Flush()
}

func (c *Context) Param(key string) string {
	return chi.URLParam(c.Request, key)
}

func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) GetString(key string) string {
	return c.Request.Context().Value(key).(string)
}
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Context) DefaultQuery(key, def string) string {
	if v := c.Query(key); v != "" {
		return v
	}
	return def
}

func (c *Context) Header(key, value string) {
	c.Writer.Header().Set(key, value)
}

type HandlerFunc func(*Context)

func Adapter(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := New(w, r)
		ctx := context.WithValue(r.Context(), ctxKey, c)
		r = r.WithContext(ctx)
		h(c)
	}
}
