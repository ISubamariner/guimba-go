"use client";

import { useCallback, useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, Badge } from "@/components/ui";
import { api } from "@/lib/api";
import type { DashboardStatsResponse, RecentActivitiesResponse } from "@/types/api";

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStatsResponse | null>(null);
  const [activities, setActivities] = useState<RecentActivitiesResponse | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    try {
      const [statsRes, activitiesRes] = await Promise.all([
        api.get<DashboardStatsResponse>("/dashboard/stats"),
        api.get<RecentActivitiesResponse>("/dashboard/recent-activities"),
      ]);
      setStats(statsRes);
      setActivities(activitiesRes);
    } catch {
      // Errors handled by API client (401 → redirect)
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (loading) {
    return <div className="text-muted">Loading dashboard...</div>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">Dashboard</h1>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="Total Tenants" value={stats?.total_tenants ?? 0} />
        <StatCard label="Total Properties" value={stats?.total_properties ?? 0} />
        <StatCard
          label="Active Debts"
          value={stats?.active_debts ?? 0}
          variant="warning"
        />
        <StatCard
          label="Overdue Debts"
          value={stats?.overdue_debts ?? 0}
          variant="danger"
        />
      </div>

      {/* Recent Activities */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Activities</CardTitle>
        </CardHeader>
        <CardContent>
          {!activities?.data?.length ? (
            <p className="text-sm text-muted">No recent activities</p>
          ) : (
            <div className="space-y-3">
              {activities.data.map((activity, i) => (
                <div
                  key={`${activity.timestamp}-${activity.action}-${i}`}
                  className="flex items-start justify-between gap-4 py-2 border-b border-border last:border-0"
                >
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-foreground">{activity.description}</p>
                    <p className="text-xs text-muted mt-0.5">
                      {new Date(activity.timestamp).toLocaleString()}
                    </p>
                  </div>
                  <Badge variant="outline">{activity.action}</Badge>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function StatCard({
  label,
  value,
  variant,
}: {
  label: string;
  value: number;
  variant?: "warning" | "danger";
}) {
  return (
    <Card>
      <CardContent>
        <p className="text-sm text-muted">{label}</p>
        <p
          className={`text-3xl font-bold mt-1 ${
            variant === "danger"
              ? "text-danger"
              : variant === "warning"
                ? "text-warning"
                : "text-foreground"
          }`}
        >
          {value}
        </p>
      </CardContent>
    </Card>
  );
}
