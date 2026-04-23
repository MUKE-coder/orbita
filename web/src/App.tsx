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
import CreateApp from "./pages/CreateApp";
import DatabaseDetail from "./pages/DatabaseDetail";
import CreateDatabase from "./pages/CreateDatabase";
import CronDetail from "./pages/CronDetail";
import CreateCron from "./pages/CreateCron";
import Marketplace from "./pages/Marketplace";
import AdminNodes from "./pages/AdminNodes";
import AdminOrgs from "./pages/AdminOrgs";
import AuditLogs from "./pages/AuditLogs";
import Projects from "./pages/Projects";
import NotFound from "./pages/NotFound";
import Landing from "./pages/Landing";
import Docs from "./pages/Docs";
import GitConnections from "./pages/GitConnections";

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
        <Route path="/" element={<Landing />} />
        <Route path="/docs" element={<Docs />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/reset-password" element={<ResetPassword />} />
        <Route path="/verify-email" element={<VerifyEmail />} />
        <Route path="/join" element={<JoinOrg />} />

        {/* Protected routes */}
        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/:orgSlug/projects"
          element={
            <ProtectedRoute>
              <Dashboard>
                <Projects />
              </Dashboard>
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
          path="/orgs/:orgSlug/apps/new"
          element={
            <ProtectedRoute>
              <Dashboard>
                <CreateApp />
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
          path="/orgs/:orgSlug/databases/new"
          element={
            <ProtectedRoute>
              <Dashboard>
                <CreateDatabase />
              </Dashboard>
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/:orgSlug/databases/:dbId"
          element={
            <ProtectedRoute>
              <Dashboard>
                <DatabaseDetail />
              </Dashboard>
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/:orgSlug/services"
          element={
            <ProtectedRoute>
              <Dashboard>
                <Marketplace />
              </Dashboard>
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/:orgSlug/cron-jobs/new"
          element={
            <ProtectedRoute>
              <Dashboard>
                <CreateCron />
              </Dashboard>
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/:orgSlug/cron-jobs/:cronId"
          element={
            <ProtectedRoute>
              <Dashboard>
                <CronDetail />
              </Dashboard>
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/:orgSlug/audit-logs"
          element={
            <ProtectedRoute>
              <Dashboard>
                <AuditLogs />
              </Dashboard>
            </ProtectedRoute>
          }
        />
        <Route
          path="/orgs/:orgSlug/git"
          element={
            <ProtectedRoute>
              <Dashboard>
                <GitConnections />
              </Dashboard>
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/nodes"
          element={
            <ProtectedRoute>
              <Dashboard>
                <AdminNodes />
              </Dashboard>
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/orgs"
          element={
            <ProtectedRoute>
              <Dashboard>
                <AdminOrgs />
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
        {/* 404 */}
        <Route path="*" element={<NotFound />} />
      </Routes>
      <Toaster richColors position="top-right" />
    </>
  );
}

export default App;
