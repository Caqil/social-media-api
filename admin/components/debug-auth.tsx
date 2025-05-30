// components/debug-auth.tsx - Debug Auth Status
"use client";

import { useAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useState } from "react";

export function DebugAuth() {
  const { user, isAuthenticated, hasValidToken, isLoading } = useAuth();
  const [debugInfo, setDebugInfo] = useState<any>(null);

  const checkTokenStatus = () => {
    const token = apiClient.getCurrentToken();
    const storedUser = apiClient.getCurrentUser();

    const debugData = {
      hasToken: !!token,
      tokenLength: token?.length || 0,
      tokenPreview: token ? `${token.substring(0, 20)}...` : "None",
      hasStoredUser: !!storedUser,
      storedUserRole: storedUser?.role || "None",
      contextAuthenticated: isAuthenticated,
      contextHasValidToken: hasValidToken,
      contextLoading: isLoading,
      contextUserRole: user?.role || "None",
      localStorage: {
        adminToken: localStorage.getItem("admin_token") ? "Present" : "Missing",
        adminRefreshToken: localStorage.getItem("admin_refresh_token")
          ? "Present"
          : "Missing",
        adminUser: localStorage.getItem("admin_user") ? "Present" : "Missing",
      },
    };

    setDebugInfo(debugData);
    console.log("🔍 Debug Info:", debugData);
  };

  const testApiCall = async () => {
    try {
      console.log("🧪 Testing API call...");
      const response = await apiClient.getDashboardStats();
      console.log("✅ API call successful:", response);
    } catch (error) {
      console.error("❌ API call failed:", error);
    }
  };

  if (process.env.NODE_ENV !== "development") {
    return null; // Only show in development
  }

  return (
    <Card className="mb-4 border-dashed border-orange-200 bg-orange-50">
      <CardHeader>
        <CardTitle className="text-sm font-medium text-orange-800">
          🔍 Debug: Auth Status
        </CardTitle>
        <CardDescription className="text-orange-600">
          Development only - Auth debugging information
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        <div className="grid grid-cols-2 gap-2 text-sm">
          <div>Context Authenticated:</div>
          <Badge variant={isAuthenticated ? "default" : "secondary"}>
            {isAuthenticated ? "Yes" : "No"}
          </Badge>

          <div>Context Has Valid Token:</div>
          <Badge variant={hasValidToken ? "default" : "secondary"}>
            {hasValidToken ? "Yes" : "No"}
          </Badge>

          <div>Context Loading:</div>
          <Badge variant={isLoading ? "secondary" : "default"}>
            {isLoading ? "Yes" : "No"}
          </Badge>

          <div>User Role:</div>
          <Badge variant="outline">{user?.role || "None"}</Badge>
        </div>

        {debugInfo && (
          <div className="mt-4 space-y-2 text-xs">
            <div className="grid grid-cols-2 gap-1">
              <div>Token Present:</div>
              <span
                className={
                  debugInfo.hasToken ? "text-green-600" : "text-red-600"
                }
              >
                {debugInfo.hasToken ? "Yes" : "No"}
              </span>

              <div>Token Length:</div>
              <span>{debugInfo.tokenLength}</span>

              <div>Stored User:</div>
              <span
                className={
                  debugInfo.hasStoredUser ? "text-green-600" : "text-red-600"
                }
              >
                {debugInfo.hasStoredUser ? "Yes" : "No"}
              </span>

              <div>LocalStorage Token:</div>
              <span
                className={
                  debugInfo.localStorage.adminToken === "Present"
                    ? "text-green-600"
                    : "text-red-600"
                }
              >
                {debugInfo.localStorage.adminToken}
              </span>

              <div>LocalStorage User:</div>
              <span
                className={
                  debugInfo.localStorage.adminUser === "Present"
                    ? "text-green-600"
                    : "text-red-600"
                }
              >
                {debugInfo.localStorage.adminUser}
              </span>
            </div>

            {debugInfo.tokenPreview && (
              <div className="mt-2">
                <div className="font-medium">Token Preview:</div>
                <code className="text-xs bg-gray-100 p-1 rounded">
                  {debugInfo.tokenPreview}
                </code>
              </div>
            )}
          </div>
        )}

        <div className="flex gap-2 mt-4">
          <Button onClick={checkTokenStatus} size="sm" variant="outline">
            Check Status
          </Button>
          <Button onClick={testApiCall} size="sm" variant="outline">
            Test API Call
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
