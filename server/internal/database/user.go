package database

func (d *DB) CreateUser(u *User) (*User, error) {
	_, err := d.db.Exec(
		"INSERT INTO users (id, email, name, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		u.ID, u.Email, u.Name, u.Password, u.CreatedAt, u.UpdatedAt,
	)
	return u, err
}

func (d *DB) GetUserByEmail(email string) (*User, error) {
	var u User
	err := d.db.QueryRow(
		"SELECT id, email, name, password, created_at, updated_at FROM users WHERE email = ?", email,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

func (d *DB) GetUserByID(id string) (*User, error) {
	var u User
	err := d.db.QueryRow(
		"SELECT id, email, name, password, created_at, updated_at FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}
