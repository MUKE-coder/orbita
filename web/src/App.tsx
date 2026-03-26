import { Routes, Route, Navigate } from "react-router-dom";
import { Toaster } from "@/components/ui/sonner";
import { useAuthStore } from "@/stores/auth";

import Login from "./pages/Login";
import Register from "./pages/Register";
import ForgotPassword from "./pages/ForgotPassword";
import ResetPassword from "./pages/ResetPassword";
import VerifyEmail from "./pages/VerifyEmail";
import Dashboard from "./pages/Dashboard";
import CreateOrg from "./pages/CreateOrg";
import OrgMembers from "./pages/OrgMembers";
import JoinOrg from "./pages/JoinOrg";
import OrgSettings from "./pages/OrgSettings";
import ProjectDetail from "./pages/ProjectDetail";
import AppDetail from "./pages/AppDetail";

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}

function App() {
  return (
    <>
      <Routes>
        {/* Public routes */}
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/reset-password" element={<ResetPassword />} />
        <Route path="/verify-email" element={<VerifyEmail />} />
        <Route path="/join" element={<JoinOrg />} />

        {/* Protected routes */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/new"
          element={
            <ProtectedRoute>
              <CreateOrg />
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/:orgSlug/projects/:projectId"
          element={
            <ProtectedRoute>
              <Dashboard>
                <ProjectDetail />
              </Dashboard>
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/:orgSlug/apps/:appId"
          element={
            <ProtectedRoute>
              <Dashboard>
                <AppDetail />
              </Dashboard>
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/:orgSlug/settings/members"
          element={
            <ProtectedRoute>
              <OrgSettings>
                <OrgMembers />
              </OrgSettings>
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/:orgSlug/settings"
          element={
            <ProtectedRoute>
              <OrgSettings />
            </ProtectedRoute>
          }
        />
      </Routes>
      <Toaster richColors position="top-right" />
    </>
  );
}

export default App;
