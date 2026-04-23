import NotificationBell from "@/components/layout/NotificationBell";
import ThemeToggle from "@/components/layout/ThemeToggle";
import { useOrgStore } from "@/stores/org";

interface TopbarProps {
  title?: string;
  description?: string;
  actions?: React.ReactNode;
}

export function Topbar({ title, description, actions }: TopbarProps) {
  const currentOrg = useOrgStore((s) => s.currentOrg);

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
      </div>
    </header>
  );
}

export default Topbar;
