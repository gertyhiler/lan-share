package pathutil

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var unsafeChars = regexp.MustCompile(`[^A-Za-z0-9._ -]+`)

// SafeFilename mirrors Python _safe_filename: basename, strip dots/spaces, sanitize.
func SafeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	name = strings.Trim(name, ".")
	name = unsafeChars.ReplaceAllString(name, "_")
	name = strings.Join(strings.Fields(name), " ")
	name = strings.TrimSpace(name)
	if name == "" {
		return "upload-" + strconv.FormatInt(time.Now().Unix(), 10)
	}
	return name
}
