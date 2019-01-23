package engine

const (
	prefixMime     = "x_lens_mime_type:"
	prefixCategory = "x_lens_category:"
)

func mimeType(t string) string { return prefixMime + t }
func category(t string) string { return prefixCategory + t }
