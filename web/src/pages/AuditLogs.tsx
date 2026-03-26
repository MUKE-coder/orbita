import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Loader2, Shield } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { notificationsApi } from "@/api/notifications";
import { useOrgStore } from "@/stores/org";

function AuditLogs() {
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const slug = currentOrg?.slug || "";
  const [page, setPage] = useState(1);

  const { data, isLoading } = useQuery({
    queryKey: ["audit-logs", slug, page],
    queryFn: () => notificationsApi.listAuditLogs(slug, page, 25),
    enabled: !!slug,
  });

  const logs = data?.data?.data?.audit_logs || [];
  const total = data?.data?.data?.total || 0;
  const totalPages = Math.ceil(total / 25);

  if (!currentOrg) {
    return <p className="text-muted-foreground text-center py-12">Select an organization.</p>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Shield className="h-5 w-5 text-muted-foreground" />
        <h2 className="text-xl font-semibold">Audit Log</h2>
        <Badge variant="secondary" className="text-xs">{total} entries</Badge>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Time</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Resource</TableHead>
                <TableHead>IP</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((log) => (
                <TableRow key={log.id}>
                  <TableCell className="text-xs">
                    {new Date(log.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary" className="text-xs">
                      {log.action}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {log.resource_type || "—"}
                    {log.resource_id && (
                      <span className="font-mono ml-1">
                        {log.resource_id.slice(0, 8)}
                      </span>
                    )}
                  </TableCell>
                  <TableCell className="text-xs font-mono text-muted-foreground">
                    {log.ip || "—"}
                  </TableCell>
                </TableRow>
              ))}
              {logs.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-muted-foreground py-8">
                    No audit log entries yet.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>

          {totalPages > 1 && (
            <div className="flex justify-center gap-2">
              <Button
                size="sm"
                variant="outline"
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
              >
                Previous
              </Button>
              <span className="text-sm text-muted-foreground flex items-center">
                Page {page} of {totalPages}
              </span>
              <Button
                size="sm"
                variant="outline"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

export default AuditLogs;
