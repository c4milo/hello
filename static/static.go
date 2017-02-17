package static

import (
	"bytes"
	"net/http"
	"strings"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/golang/glog"
)

//go:generate make embed

// Handler serves static assets and does not forward the request to any handlers. This handler is supposed to be
// initiliazed last in the middleware chain.
func Handler(h http.Handler) http.Handler {
	custom404, _ := Asset("404.html")

	uiAssets := &assetfs.AssetFS{
		Asset:     Asset,
		AssetDir:  AssetDir,
		AssetInfo: AssetInfo,
	}

	uiFileServer := http.FileServer(uiAssets)

	indexHTML, err := Asset("index.html")
	if err != nil {
		glog.Fatalf("Error finding index.html on embedded file: %v", err)
	}

	fileInfo, err := AssetInfo("index.html")
	if err != nil {
		glog.Fatalf("Error finding index.html metadata on embedded file: %+v", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For user agents accepting HTML, we serve our Single Page App.
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "html") {
			if strings.HasPrefix(r.URL.Path, "/debug") {
				h.ServeHTTP(w, r)
				return
			}

			glog.V(3).Infof("===> Serving single page app")
			http.ServeContent(w, r, "index.html", fileInfo.ModTime(), bytes.NewReader(indexHTML))
			return
		}

		glog.V(3).Infof("===> Serving static asset...")
		// Otherwise we serve static assets such as scripts, fonts, styles, images or a custom 404 page.
		rw := NewResponseWriter404(w, custom404)
		uiFileServer.ServeHTTP(rw, r)
	})
}
