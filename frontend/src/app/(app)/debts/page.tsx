"use client";

import { useState, type FormEvent } from "react";
import {
  Button,
  Input,
  Select,
  Textarea,
  Badge,
  Modal,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
  useToast,
} from "@/components/ui";
import { api, ApiClientError } from "@/lib/api";
import { useFetch } from "@/hooks/use-fetch";
import { formatMoney, formatDate } from "@/lib/format";
import type {
  PaginatedResponse,
  DebtResponse,
  DebtStatus,
  CreateDebtRequest,
  RecordPaymentRequest,
  CancelDebtRequest,
  TenantResponse,
} from "@/types/api";

const debtTypes = [
  { value: "RENT", label: "Rent" },
  { value: "UTILITIES", label: "Utilities" },
  { value: "MAINTENANCE", label: "Maintenance" },
  { value: "PENALTY", label: "Penalty" },
  { value: "OTHER", label: "Other" },
];

const paymentMethods = [
  { value: "CASH", label: "Cash" },
  { value: "BANK_TRANSFER", label: "Bank Transfer" },
  { value: "MOBILE_MONEY", label: "Mobile Money" },
  { value: "CHECK", label: "Check" },
  { value: "CREDIT_CARD", label: "Credit Card" },
  { value: "OTHER", label: "Other" },
];

const statusVariant: Record<DebtStatus, "default" | "success" | "warning" | "danger" | "outline"> = {
  PENDING: "default",
  PARTIAL: "warning",
  PAID: "success",
  OVERDUE: "danger",
  CANCELLED: "outline",
};

export default function DebtsPage() {
  const [showCreate, setShowCreate] = useState(false);
  const [payDebt, setPayDebt] = useState<DebtResponse | null>(null);
  const [cancelDebt, setCancelDebt] = useState<DebtResponse | null>(null);
  const { data, loading, refetch } = useFetch<PaginatedResponse<DebtResponse>>("/debts");
  const { toast } = useToast();

  const debts = data?.data ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">Debts</h1>
        <Button onClick={() => setShowCreate(true)}>Add Debt</Button>
      </div>

      {loading ? (
        <p className="text-muted">Loading debts...</p>
      ) : debts.length === 0 ? (
        <p className="text-muted">No debts yet.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Type</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Original</TableHead>
              <TableHead>Balance</TableHead>
              <TableHead>Due Date</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {debts.map((debt) => (
              <TableRow key={debt.id}>
                <TableCell>{debt.debt_type}</TableCell>
                <TableCell className="font-medium max-w-[200px] truncate">
                  {debt.description}
                </TableCell>
                <TableCell>{formatMoney(debt.original_amount)}</TableCell>
                <TableCell>{formatMoney(debt.balance)}</TableCell>
                <TableCell>{formatDate(debt.due_date)}</TableCell>
                <TableCell>
                  <Badge variant={statusVariant[debt.status]}>{debt.status}</Badge>
                </TableCell>
                <TableCell>
                  <div className="flex gap-1">
                    {(debt.status === "PENDING" || debt.status === "PARTIAL" || debt.status === "OVERDUE") && (
                      <Button size="sm" variant="outline" onClick={() => setPayDebt(debt)}>
                        Pay
                      </Button>
                    )}
                    {debt.status !== "PAID" && debt.status !== "CANCELLED" && (
                      <Button size="sm" variant="ghost" onClick={() => setCancelDebt(debt)}>
                        Cancel
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <CreateDebtModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={() => {
          setShowCreate(false);
          refetch();
          toast("Debt created successfully", "success");
        }}
      />

      {payDebt && (
        <PayDebtModal
          debt={payDebt}
          open={!!payDebt}
          onClose={() => setPayDebt(null)}
          onPaid={() => {
            setPayDebt(null);
            refetch();
            toast("Payment recorded successfully", "success");
          }}
        />
      )}

      {cancelDebt && (
        <CancelDebtModal
          debt={cancelDebt}
          open={!!cancelDebt}
          onClose={() => setCancelDebt(null)}
          onCancelled={() => {
            setCancelDebt(null);
            refetch();
            toast("Debt cancelled", "success");
          }}
        />
      )}
    </div>
  );
}

function CreateDebtModal({
  open,
  onClose,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
}) {
  const { data: tenantsData } = useFetch<PaginatedResponse<TenantResponse>>(open ? "/tenants" : null);
  const [tenantId, setTenantId] = useState("");
  const [debtType, setDebtType] = useState("RENT");
  const [description, setDescription] = useState("");
  const [amount, setAmount] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [notes, setNotes] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const tenantOptions = (tenantsData?.data ?? []).map((t) => ({
    value: t.id,
    label: t.full_name,
  }));

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    const body: CreateDebtRequest = {
      tenant_id: tenantId,
      debt_type: debtType as CreateDebtRequest["debt_type"],
      description,
      original_amount: { amount, currency: "PHP" },
      due_date: dueDate,
      ...(notes && { notes }),
    };

    try {
      await api.post("/debts", body);
      setTenantId("");
      setDebtType("RENT");
      setDescription("");
      setAmount("");
      setDueDate("");
      setNotes("");
      onCreated();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : "Failed to create debt");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add Debt">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
            {error}
          </div>
        )}
        <Select
          id="debt-tenant"
          label="Tenant"
          value={tenantId}
          onChange={(e) => setTenantId(e.target.value)}
          options={tenantOptions}
          placeholder="Select a tenant"
          required
        />
        <Select
          id="debt-type"
          label="Debt Type"
          value={debtType}
          onChange={(e) => setDebtType(e.target.value)}
          options={debtTypes}
        />
        <Input
          id="debt-desc"
          label="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          required
        />
        <Input
          id="debt-amount"
          label="Amount (PHP)"
          type="number"
          step="0.01"
          min="0.01"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          required
        />
        <Input
          id="debt-due"
          label="Due Date"
          type="date"
          value={dueDate}
          onChange={(e) => setDueDate(e.target.value)}
          required
        />
        <Textarea
          id="debt-notes"
          label="Notes"
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
        />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Creating..." : "Create Debt"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}

function PayDebtModal({
  debt,
  open,
  onClose,
  onPaid,
}: {
  debt: DebtResponse;
  open: boolean;
  onClose: () => void;
  onPaid: () => void;
}) {
  const [amount, setAmount] = useState(debt.balance.amount);
  const [paymentMethod, setPaymentMethod] = useState("CASH");
  const [transactionDate, setTransactionDate] = useState(
    new Date().toISOString().split("T")[0],
  );
  const [description, setDescription] = useState("");
  const [referenceNumber, setReferenceNumber] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    const body: RecordPaymentRequest = {
      debt_id: debt.id,
      tenant_id: debt.tenant_id,
      amount: { amount, currency: debt.balance.currency },
      payment_method: paymentMethod,
      transaction_date: transactionDate,
      description: description || `Payment for ${debt.description}`,
      ...(referenceNumber && { reference_number: referenceNumber }),
    };

    try {
      await api.post("/transactions/payment", body);
      onPaid();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : "Failed to record payment");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Record Payment">
      <p className="text-sm text-muted mb-4">
        Balance: <span className="font-semibold text-foreground">{formatMoney(debt.balance)}</span>
      </p>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
            {error}
          </div>
        )}
        <Input
          id="pay-amount"
          label="Amount"
          type="number"
          step="0.01"
          min="0.01"
          max={debt.balance.amount}
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          required
        />
        <Select
          id="pay-method"
          label="Payment Method"
          value={paymentMethod}
          onChange={(e) => setPaymentMethod(e.target.value)}
          options={paymentMethods}
        />
        <Input
          id="pay-date"
          label="Transaction Date"
          type="date"
          value={transactionDate}
          onChange={(e) => setTransactionDate(e.target.value)}
          required
        />
        <Input
          id="pay-desc"
          label="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <Input
          id="pay-ref"
          label="Reference Number"
          value={referenceNumber}
          onChange={(e) => setReferenceNumber(e.target.value)}
        />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Recording..." : "Record Payment"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}

function CancelDebtModal({
  debt,
  open,
  onClose,
  onCancelled,
}: {
  debt: DebtResponse;
  open: boolean;
  onClose: () => void;
  onCancelled: () => void;
}) {
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    const body: CancelDebtRequest = { ...(reason && { reason }) };

    try {
      await api.put(`/debts/${debt.id}/cancel`, body);
      onCancelled();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : "Failed to cancel debt");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Cancel Debt">
      <p className="text-sm text-muted mb-4">
        Cancel debt: <span className="font-semibold text-foreground">{debt.description}</span> ({formatMoney(debt.original_amount)})
      </p>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
            {error}
          </div>
        )}
        <Textarea
          id="cancel-reason"
          label="Reason (optional)"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="Why is this debt being cancelled?"
        />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>Keep Debt</Button>
          <Button type="submit" variant="danger" disabled={submitting}>
            {submitting ? "Cancelling..." : "Cancel Debt"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
