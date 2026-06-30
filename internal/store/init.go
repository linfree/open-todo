package store

import "time"

func (s *Store) InitDefaults() error {
	lists, err := s.GetLists()
	if err != nil {
		return err
	}
	if len(lists) > 0 {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	defaults := []TaskList{
		{ID: "all", Name: "全部", Icon: strPtr("Inbox"), OrderNum: 0, CreatedAt: now, UpdatedAt: now},
		{ID: "today", Name: "今天", Icon: strPtr("Sun"), OrderNum: 1, CreatedAt: now, UpdatedAt: now},
		{ID: "week", Name: "最近7天", Icon: strPtr("Calendar"), OrderNum: 2, CreatedAt: now, UpdatedAt: now},
	}
	for _, l := range defaults {
		ll := l
		if err := s.SaveList(&ll); err != nil {
			return err
		}
	}
	return nil
}

func strPtr(s string) *string { return &s }
