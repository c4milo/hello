package static

import "net/http"

// ResponseWriter404 wraps Go's http.ResponseWriter to check whether the status code about to be sent is 404.
// If so, it will take over the response and send back a custom body.
// We need to do this because Go does not offer an API for sending custom pages when using Go's static file server.
// Our single page application will handle rendering 404 errors.
type ResponseWriter404 struct {
	rw           http.ResponseWriter
	customHTML   []byte
	writteStatus int
}

// NewResponseWriter404 returns a 404 response writer instance.
func NewResponseWriter404(rw http.ResponseWriter, customHTML []byte) http.ResponseWriter {
	return &ResponseWriter404{
		rw:         rw,
		customHTML: customHTML,
	}
}

// Write checks whether a 404 status code was written, if so, it renders our custom HTML data back to the user.
// Otherwise, it just forwards the original data as-is.
func (w *ResponseWriter404) Write(data []byte) (int, error) {
	if w.writteStatus == http.StatusNotFound && w.customHTML != nil {
		return w.rw.Write(w.customHTML)
	}
	return w.rw.Write(data)
}

// Header just forwards the function call to the underlined ResponseWriter.
func (w *ResponseWriter404) Header() http.Header {
	return w.rw.Header()
}

// WriteHeader is called by Go's http error helper first, before writing the response body. So,
// we keep track of the written status code, to then write our custom HTML if it is a 404.
func (w *ResponseWriter404) WriteHeader(code int) {
	if code == http.StatusNotFound {
		w.rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	w.writteStatus = code
	w.rw.WriteHeader(code)
}
