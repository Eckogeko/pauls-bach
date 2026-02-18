import { createContext, useContext, useState, useEffect, type ReactNode } from "react";
import type { User } from "@/lib/types";
import * as api from "@/lib/api";

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (username: string, pin: string) => Promise<void>;
  register: (username: string, pin: string) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
  updateBalance: (newBalance: number) => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshUser = async () => {
    try {
      const u = await api.getMe();
      setUser(u);
    } catch {
      localStorage.removeItem("token");
      setUser(null);
    }
  };

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (token) {
      refreshUser().finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, []);

  const login = async (username: string, pin: string) => {
    const res = await api.login(username, pin);
    localStorage.setItem("token", res.token);
    setUser(res.user);
  };

  const register = async (username: string, pin: string) => {
    const res = await api.register(username, pin);
    localStorage.setItem("token", res.token);
    setUser(res.user);
  };

  const logout = () => {
    localStorage.removeItem("token");
    setUser(null);
  };

  const updateBalance = (newBalance: number) => {
    setUser((prev) => (prev ? { ...prev, balance: newBalance } : prev));
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout, refreshUser, updateBalance }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
