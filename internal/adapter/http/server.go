package httpadapter

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gertyhiler/lan-share/internal/domain"
	"github.com/gertyhiler/lan-share/internal/platform/netutil"
	"github.com/gertyhiler/lan-share/internal/platform/pathutil"
	filesvc "github.com/gertyhiler/lan-share/internal/usecase/files"
	pastesvc "github.com/gertyhiler/lan-share/internal/usecase/paste"
)

//go:embed templates/index.html
var indexFS embed.FS

const (
	maxPasteBody   = 2 << 20 // 2 MiB
	maxUploadBody  = 64 << 20
	maxFilePerPart = 32 << 20
)

// Paths holds display paths for the HTML UI.
type Paths struct {
	Uploads string
	Shared  string
}

// Handler is the delivery layer: HTTP mapped to use cases.
type Handler struct {
	paste      *pastesvc.Service
	files      *filesvc.Service
	port       int
	paths      Paths
	indexTmpl  *template.Template
	errLog     *log.Logger
	serverName string
}

// NewHandler builds the HTTP adapter.
func NewHandler(paste *pastesvc.Service, files *filesvc.Service, port int, paths Paths, errLog *log.Logger, serverName string) (*Handler, error) {
	tpl, err := template.ParseFS(indexFS, "templates/index.html")
	if err != nil {
		return nil, fmt.Errorf("parse index template: %w", err)
	}
	if errLog == nil {
		errLog = log.Default()
	}
	if serverName == "" {
		serverName = "lan-share/1.0"
	}
	return &Handler{
		paste:      paste,
		files:      files,
		port:       port,
		paths:      paths,
		indexTmpl:  tpl,
		errLog:     errLog,
		serverName: serverName,
	}, nil
}

// Routes returns the application mux (Go 1.22+ patterns).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /{$}", http.HandlerFunc(h.handleIndex))
	mux.Handle("GET /", http.HandlerFunc(h.handleIndex))
	mux.Handle("GET /api/paste/latest", http.HandlerFunc(h.handlePasteLatest))
	mux.Handle("GET /api/files", http.HandlerFunc(h.handleAPIFiles))
	mux.Handle("GET /api/shared", http.HandlerFunc(h.handleAPIShared))
	mux.Handle("GET /files/{name}", http.HandlerFunc(h.handleDownloadUpload))
	mux.Handle("GET /shared/{name}", http.HandlerFunc(h.handleDownloadShared))
	mux.Handle("POST /paste", http.HandlerFunc(h.handlePaste))
	mux.Handle("POST /upload", http.HandlerFunc(h.handleUpload))

	next := http.Handler(mux)
	next = h.withServerHeader(next)
	return next
}

func (h *Handler) withServerHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", h.serverName)
		next.ServeHTTP(w, r)
	})
}

type indexData struct {
	ShareHost      string // хост для подсказки «с другого устройства» (LAN, если зашли с localhost)
	Port           int
	LatestPasteB64 string
	UploadsDir     string
	SharedDir      string
	UploadList     []indexFileRow
	SharedList     []indexFileRow
}

type indexFileRow struct {
	Name    string
	Size    string
	Updated string
	HREF    template.URL
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	latest := ""
	if txt, err := h.paste.Latest(ctx); err == nil {
		latest = txt
	} else if !errors.Is(err, domain.ErrNotFound) {
		h.errLog.Printf("index: latest paste: %v", err)
	}

	uploadRows := []indexFileRow(nil)
	if list, err := h.files.ListUploads(ctx); err != nil {
		h.errLog.Printf("index: list uploads: %v", err)
	} else {
		uploadRows = indexFileRows("/files/", list)
	}
	sharedRows := []indexFileRow(nil)
	if list, err := h.files.ListShared(ctx); err != nil {
		h.errLog.Printf("index: list shared: %v", err)
	} else {
		sharedRows = indexFileRows("/shared/", list)
	}

	shareHost := shareHostForOtherDevice(r)
	var buf bytes.Buffer
	if err := h.indexTmpl.Execute(&buf, indexData{
		ShareHost:      shareHost,
		Port:           h.port,
		LatestPasteB64: base64.StdEncoding.EncodeToString([]byte(latest)),
		UploadsDir:     h.paths.Uploads,
		SharedDir:      h.paths.Shared,
		UploadList:     uploadRows,
		SharedList:     sharedRows,
	}); err != nil {
		h.errLog.Printf("index: template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, &buf)
}

func indexFileRows(prefix string, entries []domain.FileEntry) []indexFileRow {
	rows := make([]indexFileRow, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, indexFileRow{
			Name:    e.Name,
			Size:    humanBytes(e.Bytes),
			Updated: formatMTime(e.MTime),
			HREF:    template.URL(prefix + url.PathEscape(e.Name)),
		})
	}
	return rows
}

func humanBytes(n int64) string {
	if n < 0 {
		n = 0
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for i := n / unit; i >= unit && exp < 3; i /= unit {
		div *= unit
		exp++
	}
	suffix := []string{"KB", "MB", "GB", "TB"}[exp]
	return fmt.Sprintf("%.1f %s", float64(n)/float64(div), suffix)
}

func formatMTime(unix int64) string {
	return time.Unix(unix, 0).Local().Format("02.01.2006 15:04")
}

func displayHost(r *http.Request, _ int) string {
	host := r.Host
	hostOnly := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostOnly = h
	}
	if hostOnly == "" || hostOnly == "0.0.0.0" {
		ips := netutil.LANIPv4()
		if len(ips) > 0 {
			return ips[0]
		}
		return "127.0.0.1"
	}
	return hostOnly
}

// shareHostForOtherDevice returns a host reachable from the LAN for the hint URL.
// If the page was opened via loopback, the first LAN IPv4 is preferred when available.
func shareHostForOtherDevice(r *http.Request) string {
	h := displayHost(r, 0)
	if isLocalOnlyHost(h) {
		if ips := netutil.LANIPv4(); len(ips) > 0 {
			return ips[0]
		}
	}
	return h
}

func isLocalOnlyHost(host string) bool {
	switch host {
	case "", "localhost", "::1", "[::1]":
		return true
	}
	return strings.HasPrefix(host, "127.")
}

func (h *Handler) handlePasteLatest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	text, err := h.paste.Latest(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(""))
			return
		}
		h.errLog.Printf("/api/paste/latest: %v", err)
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(text))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func (h *Handler) handleAPIFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := h.files.ListUploads(ctx)
	if err != nil {
		h.errLog.Printf("/api/files: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "list failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "files": list})
}

func (h *Handler) handleAPIShared(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := h.files.ListShared(ctx)
	if err != nil {
		h.errLog.Printf("/api/shared: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "list failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "files": list})
}

func pathNameParam(r *http.Request) (string, error) {
	raw := r.PathValue("name")
	if raw == "" {
		return "", fmt.Errorf("empty name")
	}
	return url.PathUnescape(raw)
}

func (h *Handler) handleDownloadUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name, err := pathNameParam(r)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	data, err := h.files.ReadUpload(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.errLog.Printf("/files: %v", err)
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}
	safe := pathutil.SafeFilename(name)
	sendAttachment(w, safe, data)
}

func (h *Handler) handleDownloadShared(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name, err := pathNameParam(r)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	data, err := h.files.ReadShared(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.errLog.Printf("/shared: %v", err)
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}
	safe := pathutil.SafeFilename(name)
	sendAttachment(w, safe, data)
}

func sendAttachment(w http.ResponseWriter, filename string, data []byte) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) handlePaste(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ct := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		mediaType = ""
	}

	var text string
	switch {
	case mediaType == "application/x-www-form-urlencoded" || strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
		r.Body = http.MaxBytesReader(w, r.Body, maxPasteBody)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		text = r.FormValue("text")
	case mediaType == "multipart/form-data" || strings.HasPrefix(ct, "multipart/form-data"):
		r.Body = http.MaxBytesReader(w, r.Body, maxPasteBody)
		if err := r.ParseMultipartForm(maxPasteBody); err != nil {
			http.Error(w, "bad multipart", http.StatusBadRequest)
			return
		}
		text = r.FormValue("text")
	case mediaType == "text/plain" || ct == "":
		r.Body = http.MaxBytesReader(w, r.Body, maxPasteBody)
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusInternalServerError)
			return
		}
		text = string(b)
	default:
		http.Error(w, "unsupported content-type", http.StatusUnsupportedMediaType)
		return
	}

	if err := h.paste.Save(ctx, text); err != nil {
		h.errLog.Printf("/paste: %v", err)
		http.Error(w, "write error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ct := r.Header.Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(ct)
	if mediaType != "multipart/form-data" && !strings.HasPrefix(ct, "multipart/form-data") {
		http.Error(w, "expected multipart/form-data", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBody)
	if err := r.ParseMultipartForm(maxUploadBody); err != nil {
		http.Error(w, "bad multipart", http.StatusBadRequest)
		return
	}
	if r.MultipartForm == nil {
		http.Error(w, "no parts", http.StatusBadRequest)
		return
	}

	fhs := r.MultipartForm.File["file"]
	if len(fhs) == 0 {
		http.Error(w, "no parts", http.StatusBadRequest)
		return
	}

	saved := false
	for _, fh := range fhs {
		if fh.Filename == "" {
			continue
		}
		f, err := fh.Open()
		if err != nil {
			http.Error(w, "open part", http.StatusBadRequest)
			return
		}
		data, err := io.ReadAll(io.LimitReader(f, maxFilePerPart))
		_ = f.Close()
		if err != nil {
			http.Error(w, "read part", http.StatusInternalServerError)
			return
		}
		if err := h.files.SaveUpload(ctx, fh.Filename, data); err != nil {
			h.errLog.Printf("/upload: %v", err)
			http.Error(w, "failed to save: "+pathutil.SafeFilename(fh.Filename), http.StatusInternalServerError)
			return
		}
		saved = true
	}

	if !saved {
		http.Error(w, "no files", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
