"use client";
import { createContext, useContext, useEffect, useState, ReactNode } from "react";
import { STORAGE_CONFIG } from "./config";

const KEY = STORAGE_CONFIG.TOKEN_KEY;

interface AuthContextType {
  token: string | null;
  loading: boolean;
  saveToken: (t: string) => void;
  clearToken: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const t = localStorage.getItem(KEY);
    if (t) {
      console.log("Loaded token from localStorage");
      setToken(t);
    }
    setLoading(false);
  }, []);

  function saveToken(t: string) {
    console.log("saveToken called with token:", t ? `${t.substring(0, 20)}...` : "null");
    localStorage.setItem(KEY, t);
    setToken(t);
    console.log("Token saved to state, will trigger useEffect in consumers");
  }

  function clearToken() {
    localStorage.removeItem(KEY);
    setToken(null);
  }

  return (
    <AuthContext.Provider value={{ token, loading, saveToken, clearToken }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}
