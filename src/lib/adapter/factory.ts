import type { DatabaseAdapter } from './types';
import { GoDesktopAdapter } from './GoDesktopAdapter';

let cached: DatabaseAdapter | null = null;

export function createAdapter(): DatabaseAdapter {
  if (cached) return cached;

  // 检测是否在 Go 桌面环境 (访问同一来源的 Gin 服务器)
  // PWA 模式下不走这个 adapter
  cached = new GoDesktopAdapter();
  return cached;
}

export function resetAdapter(): void {
  cached = null;
}
