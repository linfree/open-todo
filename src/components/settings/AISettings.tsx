import { useState, useEffect } from "react";
import {
  Eye,
  EyeOff,
  Check,
  X,
  Loader2,
  TestTube,
} from "lucide-react";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { cn } from "../../lib/utils";
import { aiConfigApi, AIConfig } from "../../lib/api";

export function AISettings() {
  const [config, setConfig] = useState<AIConfig>({
    enabled: false,
    base_url: "https://api.openai.com/v1",
    api_key: "",
    model: "gpt-4o",
    system_prompt: "你是一个待办事项助手，帮助用户管理任务、设置提醒和规划日程。",
  });
  const [showKey, setShowKey] = useState(false);
  const [keyDirty, setKeyDirty] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    aiConfigApi.loadConfig().then((c) => c && setConfig(c));
  }, []);

  const handleTestConnection = async () => {
    if (!config.base_url || !config.api_key) {
      setMessage({ type: "error", text: "请先填写 API 地址和密钥" });
      setTimeout(() => setMessage(null), 3000);
      return;
    }
    setIsTesting(true);
    setMessage(null);
    try {
      const res = await fetch("http://localhost:25080/api/v1/ai/status");
      const status = await res.json();
      if (status.configured) {
        setMessage({ type: "success", text: "连接成功，AI 服务可用" });
      } else {
        setMessage({ type: "error", text: "AI 未正确配置" });
      }
    } catch (e: any) {
      setMessage({ type: "error", text: `连接失败: ${e.message || "未知错误"}` });
    } finally {
      setIsTesting(false);
      setTimeout(() => setMessage(null), 3000);
    }
  };

  const handleSave = async () => {
    setIsLoading(true);
    try {
      await aiConfigApi.saveConfig(config);
      setMessage({ type: "success", text: "设置已保存" });
      setKeyDirty(false);
      setTimeout(() => setMessage(null), 2000);
    } catch (e: any) {
      console.error(e);
      const errMsg = typeof e === "string" ? e : (e.message || JSON.stringify(e));
      setMessage({ type: "error", text: `保存失败: ${errMsg}` });
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyChange = (value: string) => {
    setKeyDirty(true);
    setConfig({ ...config, api_key: value });
  };

  const handleKeyFocus = () => {
    // When user focuses the key field, clear masked value so they can type fresh
    if (!keyDirty && config.api_key.includes("*")) {
      setConfig({ ...config, api_key: "" });
    }
  };

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-lg font-semibold">AI 助手</h2>
          <p className="text-sm text-muted-foreground">配置 OpenAI 兼容的 AI 服务</p>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto pr-1 sm:pr-2 space-y-6 pb-10">
        {/* Enable toggle */}
        <div className="flex items-center justify-between p-4 bg-muted/30 rounded-lg border">
          <div>
            <h3 className="font-medium text-sm sm:text-base">启用 AI 助手</h3>
            <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">
              连接 OpenAI 兼容 API 以获得智能任务建议
            </p>
          </div>
          <input
            type="checkbox"
            checked={config.enabled}
            onChange={(e) => setConfig({ ...config, enabled: e.target.checked })}
            className="w-5 h-5 sm:w-6 sm:h-6"
          />
        </div>

        {/* Config fields — shown when enabled */}
        {config.enabled && (
          <div className="space-y-4 animate-in fade-in slide-in-from-top-2 duration-200">
            {/* API 地址 */}
            <div>
              <label className="text-sm font-medium mb-1 block">API 地址</label>
              <Input
                value={config.base_url}
                onChange={(e) => setConfig({ ...config, base_url: e.target.value })}
                placeholder="https://api.openai.com/v1"
              />
            </div>

            {/* API 密钥 */}
            <div>
              <label className="text-sm font-medium mb-1 block">API 密钥</label>
              <div className="relative">
                <Input
                  type={showKey ? "text" : "password"}
                  value={config.api_key}
                  onChange={(e) => handleKeyChange(e.target.value)}
                  onFocus={handleKeyFocus}
                  placeholder="sk-..."
                />
                <button
                  type="button"
                  onClick={() => setShowKey(!showKey)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  tabIndex={-1}
                >
                  {showKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                密钥将安全存储在本地，仅在请求时发送至您配置的 API 地址
              </p>
            </div>

            {/* 模型 */}
            <div>
              <label className="text-sm font-medium mb-1 block">模型</label>
              <Input
                value={config.model}
                onChange={(e) => setConfig({ ...config, model: e.target.value })}
                placeholder="gpt-4o"
              />
            </div>

            {/* 系统提示词 */}
            <div>
              <label className="text-sm font-medium mb-1 block">系统提示词</label>
              <textarea
                value={config.system_prompt}
                onChange={(e) => setConfig({ ...config, system_prompt: e.target.value })}
                placeholder="你是一个待办事项助手..."
                rows={4}
                className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>

            {/* Save + Test buttons */}
            <div className="flex items-center gap-3 pt-2">
              <Button onClick={handleSave} disabled={isLoading} size="sm">
                保存配置
              </Button>
              <Button variant="outline" size="sm" onClick={handleTestConnection} disabled={isTesting}>
                {isTesting ? <Loader2 className="w-4 h-4 mr-1 animate-spin" /> : <TestTube className="w-4 h-4 mr-1" />}
                测试连接
              </Button>

              {message && (
                <div
                  className={cn(
                    "text-sm flex items-center gap-2",
                    message.type === "success" ? "text-green-600" : "text-red-600"
                  )}
                >
                  {message.type === "success" ? <Check className="w-4 h-4" /> : <X className="w-4 h-4" />}
                  {message.text}
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
