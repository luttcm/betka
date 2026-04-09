"use client";

import { createContext, PropsWithChildren, useContext, useEffect, useMemo, useState } from "react";

const AUTH_STORAGE_KEY = "bet_mvp_auth";

interface StoredAuth {
  token: string;
  email?: string;
}

interface AuthContextValue {
  token: string | null;
  email: string | null;
  isAuthenticated: boolean;
  signIn: (token: string, email?: string) => void;
  signOut: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const [token, setToken] = useState<string | null>(null);
  const [email, setEmail] = useState<string | null>(null);

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
    if (!raw) {
      return;
    }

    try {
      const parsed = JSON.parse(raw) as StoredAuth;
      if (parsed?.token) {
        setToken(parsed.token);
        setEmail(parsed.email ?? null);
      }
    } catch {
      window.localStorage.removeItem(AUTH_STORAGE_KEY);
    }
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      token,
      email,
      isAuthenticated: Boolean(token),
      signIn: (nextToken: string, nextEmail?: string) => {
        const normalizedEmail = nextEmail?.trim() || null;
        setToken(nextToken);
        setEmail(normalizedEmail);

        if (typeof window !== "undefined") {
          const payload: StoredAuth = {
            token: nextToken,
            email: normalizedEmail ?? undefined,
          };
          window.localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(payload));
        }
      },
      signOut: () => {
        setToken(null);
        setEmail(null);

        if (typeof window !== "undefined") {
          window.localStorage.removeItem(AUTH_STORAGE_KEY);
        }
      },
    }),
    [email, token],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }

  return context;
}

