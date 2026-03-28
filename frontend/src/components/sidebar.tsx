"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/hooks/use-auth";
import { cn } from "@/lib/cn";

interface NavItem {
  label: string;
  href: string;
  roles?: string[];
}

const navItems: NavItem[] = [
  { label: "Dashboard", href: "/dashboard" },
  { label: "Tenants", href: "/tenants", roles: ["admin", "landlord"] },
  { label: "Properties", href: "/properties", roles: ["admin", "landlord"] },
  { label: "Debts", href: "/debts", roles: ["admin", "landlord"] },
  { label: "Transactions", href: "/transactions", roles: ["admin", "landlord"] },
  { label: "Beneficiaries", href: "/beneficiaries", roles: ["admin", "staff"] },
  { label: "Programs", href: "/programs" },
  { label: "Users", href: "/users", roles: ["admin"] },
  { label: "Audit Logs", href: "/audit", roles: ["admin", "auditor"] },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user, logout, hasRole } = useAuth();

  const visibleItems = navItems.filter((item) => {
    if (!item.roles) return true;
    return item.roles.some((role) => hasRole(role));
  });

  return (
    <aside className="flex flex-col w-64 border-r border-border bg-background h-full">
      <div className="p-4 border-b border-border">
        <h1 className="text-lg font-bold text-foreground">Guimba-GO</h1>
      </div>

      <nav className="flex-1 overflow-y-auto p-2">
        {visibleItems.map((item) => {
          const isActive = pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center px-3 py-2 text-sm rounded-[var(--radius-md)] transition-colors mb-0.5",
                isActive
                  ? "bg-primary-light text-primary font-medium"
                  : "text-muted hover:bg-muted-bg hover:text-foreground",
              )}
            >
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="border-t border-border p-4">
        <div className="text-sm text-foreground font-medium truncate">
          {user?.full_name}
        </div>
        <div className="text-xs text-muted truncate">{user?.email}</div>
        <button
          onClick={logout}
          className="mt-2 text-sm text-muted hover:text-danger transition-colors"
        >
          Sign out
        </button>
      </div>
    </aside>
  );
}
