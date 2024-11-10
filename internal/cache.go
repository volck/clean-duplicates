package internal

func (s *Stats) FillCache(w Writer) map[string]bool {

	var cache = make(map[string]bool)
	allFiles := w.GetAllFiles()

	for _, f := range allFiles {
		cache[f.FilePath] = true
	}
	s.cache = &cache
	return *s.cache
}
