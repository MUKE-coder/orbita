import { useNavigate } from "react-router-dom";
import { LogOut, User } from "lucide-react";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import NotificationBell from "@/components/layout/NotificationBell";
import ThemeToggle from "@/components/layout/ThemeToggle";
import { useOrgStore } from "@/stores/org";
import { useAuthStore } from "@/stores/auth";

interface TopbarProps {
  title?: string;
  description?: string;
  actions?: React.ReactNode;
}

export function Topbar({ title, description, actions }: TopbarProps) {
  const navigate = useNavigate();
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const logout = useAuthStore((s) => s.logout);

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  return (
    <header className="flex h-14 items-center justify-between border-b border-border bg-background/80 px-6 backdrop-blur">
      <div className="flex min-w-0 flex-col">
        {title && (
          <h1 className="font-heading text-[15px] font-semibold leading-tight tracking-tight text-foreground">
            {title}
          </h1>
        )}
        {description && (
          <p className="text-xs text-muted-foreground line-clamp-1">
            {description}
          </p>
        )}
      </div>

      <div className="flex items-center gap-1">
        {actions && <div className="mr-2 flex items-center gap-2">{actions}</div>}
        {currentOrg && <NotificationBell />}
        <ThemeToggle />

        <DropdownMenu>
          <DropdownMenuTrigger
            className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/30"
            aria-label="Account"
          >
            <div className="flex h-7 w-7 items-center justify-center rounded-full bg-linear-to-br from-brand/70 to-brand text-[11px] font-semibold text-brand-foreground">
              <User className="h-3.5 w-3.5" />
            </div>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuLabel className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
              Account
            </DropdownMenuLabel>
            <DropdownMenuItem onClick={handleLogout} className="text-destructive">
              <LogOut className="mr-2 h-4 w-4" />
              Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}

export default Topbar;
