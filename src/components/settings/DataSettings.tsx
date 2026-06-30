import { useState, useEffect } from "react";
import {
  Database,
  Cloud,
  Check,
  X,
  UploadCloud,
  FolderDown,
  AlertTriangle,
} from "lucide-react";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { cn } from "../../lib/utils";
import { InputDialog } from "../ui/confirm-dialog";
import { webdavApi, WebDavSettings, isDesktop as checkIsDesktop } from "../../lib/api";

export function DataSettings() {
  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-lg font-semibold">数据同步</h2>
          <p className="text-sm text-muted-foreground">配置多端数据同步与备份</p>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto pr-1 sm:pr-2 space-y-8 pb-10">
        {/* WebDAV (Primary) */}
        <section className="space-y-4">
          <div className="flex items-center gap-2 text-primary">
            <Cloud className="w-5 h-5" />
            <h3 className="font-semibold text-base">WebDAV 云同步</h3>
          </div>
          <BackupPanel />
        </section>

        {/* Local Storage */}
        <section className="space-y-4 pt-4 border-t">
          <div className="flex items-center gap-2 text-muted-foreground">
            <Database className="w-5 h-5" />
            <h3 className="font-semibold text-base">本地存储</h3>
          </div>
          <StoragePanel />
        </section>
      </div>
    </div>
  );
}

function BackupPanel() {
  const [settings, setSettings] = useState<WebDavSettings>({
    enabled: false,
    url: "https://dav.jianguoyun.com/dav",
    username: "",
    password: "",
    basePath: "/my-todo-backups",
    autoBackup: false,
    simpleMode: true,
  });
  const [isDesktop, setIsDesktop] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [restoreInput, setRestoreInput] = useState(false);

  useEffect(() => {
    setIsDesktop(checkIsDesktop());
    webdavApi.loadSettings().then(s => s && setSettings(s));
  }, []);

  const handleSave = async () => {
    setIsLoading(true);
    try {
      await webdavApi.saveSettings(settings);
      setMessage({ type: "success", text: "设置已保存" });
      setTimeout(() => setMessage(null), 2000);
    } catch (e: any) {
      console.error(e);
      const errMsg = typeof e === 'string' ? e : (e.message || JSON.stringify(e));
      setMessage({ type: "error", text: `保存失败: ${errMsg}` });
    } finally {
      setIsLoading(false);
    }
  };

  const handleBackup = async () => {
    setIsLoading(true);
    try {
      await webdavApi.backup();
      setMessage({ type: "success", text: "备份成功" });
    } catch (e: any) {
      console.error(e);
      const errMsg = typeof e === 'string' ? e : (e.message || JSON.stringify(e));
      setMessage({ type: "error", text: `备份失败: ${errMsg}` });
    } finally {
      setIsLoading(false);
    }
  };

  const handleRestore = async () => {
    setRestoreInput(true);
  };

  const doRestore = async (filename: string) => {
    setIsLoading(true);
    try {
      await webdavApi.restore(filename);
      setMessage({ type: "success", text: "恢复成功，请重启应用" });
    } catch (e: any) {
      console.error(e);
      const errMsg = typeof e === 'string' ? e : (e.message || JSON.stringify(e));
      setMessage({ type: "error", text: `恢复失败: ${errMsg}` });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <>
    <div className="space-y-4">
      {!isDesktop && (
        <div className="text-sm p-3 rounded bg-amber-500/10 border border-amber-500/20 text-amber-700">
          仅桌面版支持完整备份功能。
        </div>
      )}

      <div className="flex items-center justify-between p-4 bg-muted/30 rounded-lg border">
        <div>
          <h3 className="font-medium text-sm sm:text-base">启用 WebDAV 服务</h3>
          <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">坚果云 / Nextcloud / NAS</p>
        </div>
        <input
          type="checkbox"
          checked={settings.enabled}
          onChange={(e) => {
            setSettings({ ...settings, enabled: e.target.checked });
            if (!e.target.checked) handleSave();
          }}
          className="w-5 h-5 sm:w-6 sm:h-6"
        />
      </div>

      {settings.enabled && (
        <div className="space-y-4 animate-in fade-in slide-in-from-top-2 duration-200">
          <div className="grid gap-4">
             <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
               <div>
                 <label className="text-sm font-medium mb-1 block">服务地址</label>
                 <Input value={settings.url} onChange={e => setSettings({...settings, url: e.target.value})} placeholder="https://..." />
               </div>
               <div>
                 <label className="text-sm font-medium mb-1 block">存储路径</label>
                 <Input value={settings.basePath} onChange={e => setSettings({...settings, basePath: e.target.value})} placeholder="/my-backups" />
               </div>
             </div>
             <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
               <div>
                 <label className="text-sm font-medium mb-1 block">用户名</label>
                 <Input 
                   value={settings.username} 
                   onChange={e => setSettings({...settings, username: e.target.value})} 
                   placeholder="注册邮箱"
                 />
               </div>
               <div>
                 <label className="text-sm font-medium mb-1 block">密码</label>
                 <Input 
                   type="password" 
                   value={settings.password} 
                   onChange={e => setSettings({...settings, password: e.target.value})} 
                   placeholder="应用密码"
                 />
                 <p className="text-xs text-muted-foreground mt-1">
                   坚果云请使用<a href="https://help.jianguoyun.com/?p=2064" target="_blank" rel="noopener noreferrer" className="underline hover:text-primary">应用密码</a>，而非登录密码。
                 </p>
               </div>
             </div>
             
             <div className="flex flex-col sm:flex-row gap-4 pt-2">
                <label className="flex items-center gap-2 text-sm p-2 border rounded hover:bg-muted/50 cursor-pointer">
                  <input type="checkbox" checked={settings.autoBackup} onChange={e => setSettings({...settings, autoBackup: e.target.checked})} />
                  <span>自动备份 (每日)</span>
                </label>
                <label className="flex items-center gap-2 text-sm p-2 border rounded hover:bg-muted/50 cursor-pointer">
                  <input type="checkbox" checked={settings.simpleMode} onChange={e => setSettings({...settings, simpleMode: e.target.checked})} />
                  <span>精简模式 (仅数据库文件)</span>
                </label>
             </div>
          </div>

          <div className="flex flex-col sm:flex-row items-center gap-3 pt-4 border-t">
            <Button onClick={handleSave} disabled={isLoading} size="sm" className="w-full sm:w-auto">保存配置</Button>
            <div className="flex gap-2 w-full sm:w-auto">
                <Button variant="outline" onClick={handleBackup} disabled={isLoading || !isDesktop} size="sm" className="flex-1 sm:flex-none">
                <UploadCloud className="w-4 h-4 mr-2" /> 立即备份
                </Button>
                <Button variant="outline" onClick={handleRestore} disabled={isLoading || !isDesktop} size="sm" className="flex-1 sm:flex-none">
                <FolderDown className="w-4 h-4 mr-2" /> 恢复备份
                </Button>
            </div>
          </div>

          {message && (
            <div className={cn("text-sm flex items-center gap-2 p-2 rounded bg-muted/50", message.type === "success" ? "text-green-600" : "text-red-600")}>
              {message.type === "success" ? <Check className="w-4 h-4" /> : <X className="w-4 h-4" />}
              {message.text}
            </div>
          )}
        </div>
      )}
    </div>
      <InputDialog
        open={restoreInput}
        onOpenChange={setRestoreInput}
        title="恢复备份"
        placeholder="输入要恢复的备份文件名"
        confirmLabel="恢复"
        onConfirm={doRestore}
      />
    </>
  );
}

function StoragePanel() {
  return (
    <div className="space-y-4">
      <div className="p-4 rounded-lg bg-muted/30 border">
        <div className="flex items-start gap-3">
          <Database className="w-5 h-5 text-muted-foreground mt-0.5 flex-shrink-0" />
          <div className="min-w-0">
            <h3 className="font-medium text-sm sm:text-base">本地 SQLite 数据库</h3>
            <p className="text-xs sm:text-sm text-muted-foreground mt-1 break-words">
              AppData/my-todo/todo.db
            </p>
          </div>
        </div>
      </div>
      
      <div className="p-4 rounded-lg border border-red-200 bg-red-50/50">
        <h3 className="font-medium text-red-800 flex items-center gap-2 text-sm sm:text-base">
          <AlertTriangle className="w-4 h-4" /> 危险区域
        </h3>
        <div className="mt-3">
            <p className="text-xs text-muted-foreground">请手动删除 ~/.open-todo/open-todo.db 以重置数据。</p>
        </div>
      </div>
    </div>
  );
}
