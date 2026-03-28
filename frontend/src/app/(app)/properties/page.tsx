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
import { formatDate } from "@/lib/format";
import type {
  PaginatedResponse,
  PropertyResponse,
  CreatePropertyRequest,
} from "@/types/api";

const propertyTypes = [
  { value: "LAND", label: "Land" },
  { value: "BUILDING", label: "Building" },
  { value: "RESIDENTIAL", label: "Residential" },
  { value: "COMMERCIAL", label: "Commercial" },
  { value: "OTHER", label: "Other" },
];

export default function PropertiesPage() {
  const [showCreate, setShowCreate] = useState(false);
  const { data, loading, refetch } = useFetch<PaginatedResponse<PropertyResponse>>("/properties");
  const { toast } = useToast();

  const properties = data?.data ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">Properties</h1>
        <Button onClick={() => setShowCreate(true)}>Add Property</Button>
      </div>

      {loading ? (
        <p className="text-muted">Loading properties...</p>
      ) : properties.length === 0 ? (
        <p className="text-muted">No properties yet. Add your first property to get started.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Code</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Size (sqm)</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {properties.map((property) => (
              <TableRow key={property.id}>
                <TableCell className="font-medium">{property.name}</TableCell>
                <TableCell>{property.property_code}</TableCell>
                <TableCell>{property.property_type}</TableCell>
                <TableCell>{property.size_in_sqm.toLocaleString()}</TableCell>
                <TableCell>
                  <Badge variant={property.is_active ? "success" : "danger"}>
                    {property.is_active ? "Active" : "Inactive"}
                  </Badge>
                </TableCell>
                <TableCell>{formatDate(property.created_at)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <CreatePropertyModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={() => {
          setShowCreate(false);
          refetch();
          toast("Property created successfully", "success");
        }}
      />
    </div>
  );
}

function CreatePropertyModal({
  open,
  onClose,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
}) {
  const [name, setName] = useState("");
  const [propertyCode, setPropertyCode] = useState("");
  const [propertyType, setPropertyType] = useState("LAND");
  const [sizeInSqm, setSizeInSqm] = useState("");
  const [monthlyRent, setMonthlyRent] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    const body: CreatePropertyRequest = {
      name,
      property_code: propertyCode,
      property_type: propertyType,
      size_in_sqm: parseFloat(sizeInSqm),
      ...(monthlyRent && { monthly_rent_amount: parseFloat(monthlyRent) }),
      ...(description && { description }),
    };

    try {
      await api.post("/properties", body);
      setName("");
      setPropertyCode("");
      setPropertyType("LAND");
      setSizeInSqm("");
      setMonthlyRent("");
      setDescription("");
      onCreated();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : "Failed to create property");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add Property">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
            {error}
          </div>
        )}
        <Input
          id="prop-name"
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
        <Input
          id="prop-code"
          label="Property Code"
          value={propertyCode}
          onChange={(e) => setPropertyCode(e.target.value)}
          required
        />
        <Select
          id="prop-type"
          label="Property Type"
          value={propertyType}
          onChange={(e) => setPropertyType(e.target.value)}
          options={propertyTypes}
        />
        <Input
          id="prop-size"
          label="Size (sqm)"
          type="number"
          step="0.01"
          min="0.01"
          value={sizeInSqm}
          onChange={(e) => setSizeInSqm(e.target.value)}
          required
        />
        <Input
          id="prop-rent"
          label="Monthly Rent Amount"
          type="number"
          step="0.01"
          min="0"
          value={monthlyRent}
          onChange={(e) => setMonthlyRent(e.target.value)}
        />
        <Textarea
          id="prop-desc"
          label="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Creating..." : "Create Property"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
