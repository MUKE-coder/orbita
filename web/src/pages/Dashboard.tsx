import { ReactNode } from "react";

import Sidebar from "@/components/layout/Sidebar";
import Topbar from "@/components/layout/Topbar";
import DashboardOverview from "./DashboardOverview";

interface DashboardProps {
  children?: ReactNode;
  title?: string;
  description?: string;
  actions?: ReactNode;
}

function Dashboard({ children, title, description, actions }: DashboardProps) {
  return (
    <div className="flex h-screen overflow-hidden bg-background">
      <Sidebar />

      <div className="flex min-w-0 flex-1 flex-col">
        <Topbar title={title} description={description} actions={actions} />

        <main className="flex-1 overflow-y-auto">
          <div className="mx-auto w-full max-w-7xl px-6 py-8">
            {children || <DashboardOverview />}
          </div>
        </main>
      </div>
    </div>
  );
}

export default Dashboard;
