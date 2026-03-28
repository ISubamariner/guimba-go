"use client";

import { useState, type FormEvent } from "react";
import {
  Button,
  Input,
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
import { formatDate } from "@/lib/format";
import type {
  PaginatedResponse,
  TenantResponse,
  CreateTenantRequest,
} from "@/types/api";

export default function TenantsPage() {
  const [showCreate, setShowCreate] = useState(false);
  const { data, loading, refetch } = useFetch<PaginatedResponse<TenantResponse>>("/tenants");
  const { toast } = useToast();

  const tenants = data?.data ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">Tenants</h1>
        <Button onClick={() => setShowCreate(true)}>Add Tenant</Button>
      </div>

      {loading ? (
        <p className="text-muted">Loading tenants...</p>
      ) : tenants.length === 0 ? (
        <p className="text-muted">No tenants yet. Add your first tenant to get started.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Phone</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {tenants.map((tenant) => (
              <TableRow key={tenant.id}>
                <TableCell className="font-medium">{tenant.full_name}</TableCell>
                <TableCell>{tenant.email ?? "—"}</TableCell>
                <TableCell>{tenant.phone_number ?? "—"}</TableCell>
                <TableCell>
                  <Badge variant={tenant.is_active ? "success" : "danger"}>
                    {tenant.is_active ? "Active" : "Inactive"}
                  </Badge>
                </TableCell>
                <TableCell>{formatDate(tenant.created_at)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <CreateTenantModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={() => {
          setShowCreate(false);
          refetch();
          toast("Tenant created successfully", "success");
        }}
      />
    </div>
  );
}

function CreateTenantModal({
  open,
  onClose,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
}) {
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [phoneNumber, setPhoneNumber] = useState("");
  const [notes, setNotes] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    const body: CreateTenantRequest = {
      full_name: fullName,
      ...(email && { email }),
      ...(phoneNumber && { phone_number: phoneNumber }),
      ...(notes && { notes }),
    };

    try {
      await api.post("/tenants", body);
      setFullName("");
      setEmail("");
      setPhoneNumber("");
      setNotes("");
      onCreated();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : "Failed to create tenant");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add Tenant">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
            {error}
          </div>
        )}
        <Input
          id="tenant-name"
          label="Full Name"
          value={fullName}
          onChange={(e) => setFullName(e.target.value)}
          required
        />
        <Input
          id="tenant-email"
          label="Email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <Input
          id="tenant-phone"
          label="Phone Number"
          value={phoneNumber}
          onChange={(e) => setPhoneNumber(e.target.value)}
        />
        <Textarea
          id="tenant-notes"
          label="Notes"
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
        />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Creating..." : "Create Tenant"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
