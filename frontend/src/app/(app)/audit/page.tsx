"use client";

import {
  Badge,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui";
import { useFetch } from "@/hooks/use-fetch";
import { formatDateTime } from "@/lib/format";
import { useAuth } from "@/hooks/use-auth";
import type { PaginatedResponse, AuditEntryResponse } from "@/types/api";

export default function AuditPage() {
  const { hasRole } = useAuth();

  const endpoint = hasRole("admin") || hasRole("auditor") ? "/audit" : "/audit/landlord";
  const { data, loading } = useFetch<PaginatedResponse<AuditEntryResponse>>(endpoint);

  const entries = data?.data ?? [];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">Audit Logs</h1>

      {loading ? (
        <p className="text-muted">Loading audit logs...</p>
      ) : entries.length === 0 ? (
        <p className="text-muted">No audit entries found.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Timestamp</TableHead>
              <TableHead>User</TableHead>
              <TableHead>Action</TableHead>
              <TableHead>Resource</TableHead>
              <TableHead>Method</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {entries.map((entry) => (
              <TableRow key={entry.id}>
                <TableCell className="whitespace-nowrap">
                  {formatDateTime(entry.timestamp)}
                </TableCell>
                <TableCell>{entry.user_email}</TableCell>
                <TableCell>{entry.action}</TableCell>
                <TableCell>
                  {entry.resource_type}
                  <span className="text-muted text-xs ml-1">#{entry.resource_id.slice(0, 8)}</span>
                </TableCell>
                <TableCell>
                  <Badge variant="outline">{entry.method}</Badge>
                </TableCell>
                <TableCell>
                  <Badge variant={entry.success ? "success" : "danger"}>
                    {entry.status_code}
                  </Badge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  );
}
