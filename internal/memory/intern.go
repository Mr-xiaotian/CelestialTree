package memory

func (s *Store) internType(t string) string {
	if existing, ok := s.typeIntern[t]; ok {
		return existing
	}
	s.typeIntern[t] = t
	return t
}
