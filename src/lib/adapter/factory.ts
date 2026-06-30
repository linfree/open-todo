import type { DatabaseAdapter } from './types';
import { GoDesktopAdapter } from './GoDesktopAdapter';
import { DexieAdapter } from './DexieAdapter';

let cached: DatabaseAdapter | null = null;

function isGoDesktop(): boolean {
  // 检测是否连接了 Go Gin 后端
  // 在桌面环境下, Go Gin 服务运行在 localhost
  const url = new URL(window.location.href);
  return url.hostname === 'localhost' || url.hostname === '127.0.0.1';
}

export function createAdapter(): DatabaseAdapter {
  if (cached) return cached;

  if (isGoDesktop()) {
    cached = new GoDesktopAdapter();
  } else {
    cached = new DexieAdapter();
  }
  return cached;
}

export function resetAdapter(): void {
  cached = null;
}
