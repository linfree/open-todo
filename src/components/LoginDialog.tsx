import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
  DialogClose,
} from "./ui/dialog";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { cn } from "../lib/utils";
import { LogIn, UserPlus, Server, Mail, Lock, AlertCircle, Eye, EyeOff } from "lucide-react";

interface LoginDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onLoginSuccess?: (user: { id: string; email: string; name: string }, token: string) => void;
}

type TabMode = "login" | "register";

export function LoginDialog({ isOpen, onClose, onLoginSuccess }: LoginDialogProps) {
  const [mode, setMode] = useState<TabMode>("login");
  const [serverUrl, setServerUrl] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setServerUrl(localStorage.getItem("server_url") || "http://localhost:8080");
      setEmail("");
      setPassword("");
      setName("");
      setError("");
      setLoading(false);
    }
  }, [isOpen]);

  function getBaseUrl(): string {
    const url = serverUrl.trim().replace(/\/+$/, "");
    return url || "http://localhost:8080";
  }

  function saveServerUrl(url: string) {
    localStorage.setItem("server_url", url);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    const baseUrl = getBaseUrl();
    saveServerUrl(baseUrl);

    const endpoint = mode === "login"
      ? `${baseUrl}/api/v1/auth/login`
      : `${baseUrl}/api/v1/auth/register`;

    const body: Record<string, string> = { email: email.trim(), password };
    if (mode === "register") {
      body.name = name.trim();
    }

    try {
      const res = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      const data = await res.json();

      if (!res.ok) {
        setError(data.error || `Request failed (${res.status})`);
        setLoading(false);
        return;
      }

      // Store token and user info
      localStorage.setItem("auth_token", data.token);
      localStorage.setItem("user_info", JSON.stringify(data.user));

      onLoginSuccess?.(data.user, data.token);
      onClose();
    } catch (err: any) {
      setError(err.message || "Network error - please check the server URL and try again");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-md">
        <DialogTitle className="sr-only">
          {mode === "login" ? "登录" : "注册"}
        </DialogTitle>
        <DialogDescription className="sr-only">
          {mode === "login" ? "使用已有账号登录" : "创建新账号"}
        </DialogDescription>

        <DialogClose onClick={onClose} />

        {/* Header with tabs */}
        <div className="flex items-center justify-between pt-2">
          <h2 className="text-lg font-semibold text-foreground">
            {mode === "login" ? "登录" : "注册"}
          </h2>
        </div>

        {/* Tab switcher */}
        <div className="flex rounded-lg bg-muted p-1">
          <button
            type="button"
            onClick={() => { setMode("login"); setError(""); }}
            className={cn(
              "flex-1 flex items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors",
              mode === "login"
                ? "bg-background text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            <LogIn className="w-4 h-4" />
            登录
          </button>
          <button
            type="button"
            onClick={() => { setMode("register"); setError(""); }}
            className={cn(
              "flex-1 flex items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors",
              mode === "register"
                ? "bg-background text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            <UserPlus className="w-4 h-4" />
            注册
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Server URL */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-muted-foreground flex items-center gap-1.5">
              <Server className="w-3.5 h-3.5" />
              服务端地址
            </label>
            <Input
              type="url"
              value={serverUrl}
              onChange={(e) => setServerUrl(e.target.value)}
              placeholder="http://localhost:8080"
              className="h-10"
              required
            />
          </div>

          {/* Email */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-muted-foreground flex items-center gap-1.5">
              <Mail className="w-3.5 h-3.5" />
              邮箱
            </label>
            <Input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="your@email.com"
              className="h-10"
              required
            />
          </div>

          {/* Name (register only) */}
          {mode === "register" && (
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground flex items-center gap-1.5">
                <UserPlus className="w-3.5 h-3.5" />
                昵称
              </label>
              <Input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="你的昵称"
                className="h-10"
              />
            </div>
          )}

          {/* Password */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-muted-foreground flex items-center gap-1.5">
              <Lock className="w-3.5 h-3.5" />
              密码
            </label>
            <div className="relative">
              <Input
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={mode === "register" ? "至少6位密码" : "输入密码"}
                className="h-10 pr-10"
                required
                minLength={mode === "register" ? 6 : undefined}
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded text-muted-foreground hover:text-foreground transition-colors"
                tabIndex={-1}
                aria-label={showPassword ? "隐藏密码" : "显示密码"}
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </div>

          {/* Error message */}
          {error && (
            <div className="flex items-start gap-2 rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
              <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Submit */}
          <Button
            type="submit"
            className="w-full h-10"
            disabled={loading}
          >
            {loading
              ? "处理中..."
              : mode === "login"
                ? "登录"
                : "注册"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
