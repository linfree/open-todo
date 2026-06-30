package store

func (s *Store) GetTags(userID string) ([]Tag, error) {
	return s.DB.GetTags(userID)
}

func (s *Store) SaveTag(t *Tag) error {
	return s.DB.SaveTag(t)
}

func (s *Store) DeleteTag(id string) error {
	return s.DB.DeleteTag(id)
}
