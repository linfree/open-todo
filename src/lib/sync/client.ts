export interface SyncRecord {
  table_name: string;
  record_id: string;
  action: string;
  timestamp: string;
}

interface PullResponse {
  changes: SyncRecord[];
  server_time: string;
}

export class SyncClient {
  constructor(private serverUrl: string, private token: string) {}

  async push(changes: SyncRecord[]): Promise<void> {
    const res = await fetch(`${this.serverUrl}/api/v1/sync/push`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${this.token}`,
      },
      body: JSON.stringify({ changes }),
    });
    if (!res.ok) throw new Error(`push: ${res.status}`);
  }

  async pull(since: string): Promise<PullResponse> {
    const res = await fetch(`${this.serverUrl}/api/v1/sync/pull`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${this.token}`,
      },
      body: JSON.stringify({ since }),
    });
    if (!res.ok) throw new Error(`pull: ${res.status}`);
    return res.json();
  }
}
