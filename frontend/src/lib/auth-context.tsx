"use client";

import { createContext, PropsWithChildren, useContext, useEffect, useMemo, useState } from "react";

import { UserRole } from "@/lib/types";

const AUTH_STORAGE_KEY = "bet_mvp_auth";

interface StoredAuth {
  token: string;
  email?: string;
  role?: UserRole;
}

interface AuthContextValue {
  token: string | null;
  userId: string | null;
  email: string | null;
  role: UserRole | null;
  isAdmin: boolean;
  canModerate: boolean;
  isAuthenticated: boolean;
  signIn: (token: string, email?: string, role?: UserRole) => void;
  signOut: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

function roleFromToken(token: string): UserRole | null {
  try {
    const [, payload] = token.split(".");
    if (!payload) {
      return null;
    }

    const normalized = payload.replace(/-/g, "+").replace(/_/g, "/");
    const withPadding = normalized + "=".repeat((4 - (normalized.length % 4)) % 4);
    const decoded = JSON.parse(atob(withPadding)) as { role?: unknown };

    if (decoded.role === "user" || decoded.role === "moderator" || decoded.role === "admin") {
      return decoded.role;
    }
  } catch {
    return null;
  }

  return null;
}

function userIdFromToken(token: string): string | null {
  try {
    const [, payload] = token.split(".");
    if (!payload) {
      return null;
    }

    const normalized = payload.replace(/-/g, "+").replace(/_/g, "/");
    const withPadding = normalized + "=".repeat((4 - (normalized.length % 4)) % 4);
    const decoded = JSON.parse(atob(withPadding)) as { sub?: unknown };
    return typeof decoded.sub === "string" ? decoded.sub : null;
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: PropsWithChildren) {
  const [token, setToken] = useState<string | null>(null);
  const [userId, setUserId] = useState<string | null>(null);
  const [email, setEmail] = useState<string | null>(null);
  const [role, setRole] = useState<UserRole | null>(null);

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
        setUserId(userIdFromToken(parsed.token));
        setEmail(parsed.email ?? null);
        setRole(parsed.role ?? roleFromToken(parsed.token));
      }
    } catch {
      window.localStorage.removeItem(AUTH_STORAGE_KEY);
    }
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      token,
      userId,
      email,
      role,
      isAdmin: role === "admin",
      canModerate: role === "admin" || role === "moderator",
      isAuthenticated: Boolean(token),
      signIn: (nextToken: string, nextEmail?: string, nextRole?: UserRole) => {
        const normalizedEmail = nextEmail?.trim() || null;
        const resolvedRole = nextRole ?? roleFromToken(nextToken);

        setToken(nextToken);
        setUserId(userIdFromToken(nextToken));
        setEmail(normalizedEmail);
        setRole(resolvedRole);

        if (typeof window !== "undefined") {
          const payload: StoredAuth = {
            token: nextToken,
            email: normalizedEmail ?? undefined,
            role: resolvedRole ?? undefined,
          };
          window.localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(payload));
        }
      },
      signOut: () => {
        setToken(null);
        setUserId(null);
        setEmail(null);
        setRole(null);

        if (typeof window !== "undefined") {
          window.localStorage.removeItem(AUTH_STORAGE_KEY);
        }
      },
    }),
    [email, role, token, userId],
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
