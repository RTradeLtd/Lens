package utils

// Unique is used to remove duplicates from a slice
func Unique(data []string) []string {
	d := []string{}
	found := make(map[string]bool)
	for _, v := range data {
		if found[v] {
			continue
		}
		found[v] = true
		d = append(d, v)
	}
	return d
}
