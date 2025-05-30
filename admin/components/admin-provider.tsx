"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { useWebSocket } from "@/hooks/use-websocket";
import { useAuth } from "@/contexts/auth-context";

interface AdminContextType {
  realtimeData: any;
  systemAlerts: any[];
  isRealtimeConnected: boolean;
  connectionError: string | null;
}

const AdminContext = createContext<AdminContextType | undefined>(undefined);

export function AdminProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  const [realtimeData, setRealtimeData] = useState<any>(null);
  const [systemAlerts, setSystemAlerts] = useState<any[]>([]);

  const { isConnected, error } = useWebSocket("dashboard", {
    onMessage: (data) => {
      if (data.type === "stats") {
        setRealtimeData(data.payload);
      } else if (data.type === "alert") {
        setSystemAlerts((prev) => [data.payload, ...prev].slice(0, 10));
      }
    },
  });

  const value = {
    realtimeData,
    systemAlerts,
    isRealtimeConnected: isConnected && isAuthenticated,
    connectionError: error,
  };

  return (
    <AdminContext.Provider value={value}>{children}</AdminContext.Provider>
  );
}

export function useAdmin() {
  const context = useContext(AdminContext);
  if (context === undefined) {
    throw new Error("useAdmin must be used within an AdminProvider");
  }
  return context;
}
