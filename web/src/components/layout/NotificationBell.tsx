import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Bell, Check } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { notificationsApi } from "@/api/notifications";
import { useOrgStore } from "@/stores/org";

function NotificationBell() {
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const queryClient = useQueryClient();
  const slug = currentOrg?.slug || "";

  const { data } = useQuery({
    queryKey: ["notifications", slug],
    queryFn: () => notificationsApi.list(slug),
    enabled: !!slug,
    refetchInterval: 15000,
  });

  const markAllMut = useMutation({
    mutationFn: () => notificationsApi.markAllRead(slug),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["notifications", slug] }),
  });

  const notifs = data?.data?.data?.notifications || [];
  const unread = data?.data?.data?.unread_count || 0;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="relative inline-flex items-center justify-center rounded-md p-2 hover:bg-accent">
        <Bell className="h-5 w-5" />
        {unread > 0 && (
          <span className="absolute -top-0.5 -right-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-destructive text-[10px] font-bold text-white">
            {unread > 9 ? "9+" : unread}
          </span>
        )}
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-80" align="end">
        <div className="flex items-center justify-between px-3 py-2">
          <span className="text-sm font-medium">Notifications</span>
          {unread > 0 && (
            <Button
              variant="ghost"
              size="sm"
              className="text-xs h-6"
              onClick={() => markAllMut.mutate()}
            >
              <Check className="mr-1 h-3 w-3" />
              Mark all read
            </Button>
          )}
        </div>
        <DropdownMenuSeparator />
        {notifs.length === 0 ? (
          <div className="px-3 py-6 text-center text-sm text-muted-foreground">
            No notifications
          </div>
        ) : (
          notifs.slice(0, 10).map((n) => (
            <DropdownMenuItem key={n.id} className="flex flex-col items-start gap-0.5 px-3 py-2">
              <div className="flex items-center gap-2 w-full">
                {!n.read && (
                  <span className="h-2 w-2 rounded-full bg-primary shrink-0" />
                )}
                <span className="text-sm font-medium truncate">{n.title}</span>
              </div>
              {n.body && (
                <span className="text-xs text-muted-foreground line-clamp-1">
                  {n.body}
                </span>
              )}
              <span className="text-[10px] text-muted-foreground">
                {new Date(n.created_at).toLocaleString()}
              </span>
            </DropdownMenuItem>
          ))
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default NotificationBell;
