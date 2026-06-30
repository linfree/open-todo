package app

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/linfree/open-todo/internal/store"
)

type WebDAVConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func BackupToWebDAV(st *store.Store, cfg WebDAVConfig) (string, error) {
	tasks, _ := st.GetTasks()
	lists, _ := st.GetLists()

	data := map[string]interface{}{
		"tasks":     tasks,
		"lists":     lists,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("open-todo-backup.json")
	f.Write(jsonData)
	w.Close()

	filename := fmt.Sprintf("open-todo-%s.zip", time.Now().UTC().Format("20060102-150405"))
	url := fmt.Sprintf("%s/%s", cfg.URL, filename)

	req, err := http.NewRequest("PUT", url, &buf)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(cfg.Username, cfg.Password)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("webdav error: %s", resp.Status)
	}

	return filename, nil
}

func RestoreFromWebDAV(st *store.Store, cfg WebDAVConfig, filename string) error {
	url := fmt.Sprintf("%s/%s", cfg.URL, filename)
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(cfg.Username, cfg.Password)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	zr, _ := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))

	for _, f := range zr.File {
		if f.Name == "open-todo-backup.json" {
			rc, _ := f.Open()
			defer rc.Close()
			data, _ := io.ReadAll(rc)

			var backup struct {
				Tasks []store.Task     `json:"tasks"`
				Lists []store.TaskList `json:"lists"`
			}
			json.Unmarshal(data, &backup)

			for _, t := range backup.Tasks {
				tt := t
				st.SaveTask(&tt)
			}
			for _, l := range backup.Lists {
				ll := l
				st.SaveList(&ll)
			}
		}
	}
	return nil
}
