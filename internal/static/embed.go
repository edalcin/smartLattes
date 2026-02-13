package static

import "embed"

//go:embed *.html *.png *.ico css js images
var Files embed.FS
