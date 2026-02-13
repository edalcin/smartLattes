package static

import "embed"

//go:embed *.html *.png css js images
var Files embed.FS
