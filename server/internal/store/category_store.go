package store

func (s *Store) GetCategories(userID string) ([]Category, error) {
	return s.DB.GetCategories(userID)
}

func (s *Store) SaveCategory(c *Category) error {
	return s.DB.SaveCategory(c)
}

func (s *Store) DeleteCategory(id string) error {
	return s.DB.DeleteCategory(id)
}
