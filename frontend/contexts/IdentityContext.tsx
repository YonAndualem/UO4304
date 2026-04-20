"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { loadIdentity, saveIdentity, clearIdentity } from "@/lib/identity";
import type { Identity } from "@/lib/types";

interface IdentityContextValue {
  identity: Identity | null;
  setIdentity: (id: Identity) => void;
  logout: () => void;
}

const IdentityContext = createContext<IdentityContextValue | null>(null);

export function IdentityProvider({ children }: { children: ReactNode }) {
  const [identity, setIdentityState] = useState<Identity | null>(null);

  useEffect(() => {
    setIdentityState(loadIdentity());
  }, []);

  function setIdentity(id: Identity) {
    saveIdentity(id);
    setIdentityState(id);
  }

  function logout() {
    clearIdentity();
    setIdentityState(null);
  }

  return (
    <IdentityContext.Provider value={{ identity, setIdentity, logout }}>
      {children}
    </IdentityContext.Provider>
  );
}

export function useIdentity(): IdentityContextValue {
  const ctx = useContext(IdentityContext);
  if (!ctx) throw new Error("useIdentity must be used inside IdentityProvider");
  return ctx;
}
