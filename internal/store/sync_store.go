package store

type ChangeRecord struct {
	ID        int64  `json:"id"`
	TableName string `json:"table_name"`
	RecordID  string `json:"record_id"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	Synced    bool   `json:"synced"`
}

func (s *Store) GetUnsyncedChanges() ([]ChangeRecord, error) {
	rows, err := s.db.Query("SELECT id, table_name, record_id, action, timestamp FROM sync_log WHERE synced = 0 ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []ChangeRecord
	for rows.Next() {
		var c ChangeRecord
		var synced int
		if err := rows.Scan(&c.ID, &c.TableName, &c.RecordID, &c.Action, &c.Timestamp); err != nil {
			return nil, err
		}
		c.Synced = synced != 0
		changes = append(changes, c)
	}
	if changes == nil {
		changes = []ChangeRecord{}
	}
	return changes, rows.Err()
}

func (s *Store) MarkChangesSynced(ids []int64) error {
	for _, id := range ids {
		if _, err := s.db.Exec("UPDATE sync_log SET synced = 1 WHERE id = ?", id); err != nil {
			return err
		}
	}
	return nil
}
