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
import { formatMoney, formatDate } from "@/lib/format";
import type { PaginatedResponse, TransactionResponse } from "@/types/api";

export default function TransactionsPage() {
  const { data, loading } = useFetch<PaginatedResponse<TransactionResponse>>("/transactions");

  const transactions = data?.data ?? [];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">Transactions</h1>

      {loading ? (
        <p className="text-muted">Loading transactions...</p>
      ) : transactions.length === 0 ? (
        <p className="text-muted">No transactions yet.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Date</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Amount</TableHead>
              <TableHead>Method</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Reference</TableHead>
              <TableHead>Verified</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {transactions.map((tx) => (
              <TableRow key={tx.id}>
                <TableCell>{formatDate(tx.transaction_date)}</TableCell>
                <TableCell>
                  <Badge variant={tx.transaction_type === "PAYMENT" ? "success" : "warning"}>
                    {tx.transaction_type}
                  </Badge>
                </TableCell>
                <TableCell className="font-medium">{formatMoney(tx.amount)}</TableCell>
                <TableCell>{tx.payment_method}</TableCell>
                <TableCell className="max-w-[200px] truncate">{tx.description}</TableCell>
                <TableCell>{tx.reference_number ?? "—"}</TableCell>
                <TableCell>
                  <Badge variant={tx.is_verified ? "success" : "outline"}>
                    {tx.is_verified ? "Verified" : "Pending"}
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
