// API Types - matching backend Go DTOs

// ==================== Error Types ====================

export type ErrorCode =
  | "NOT_FOUND"
  | "VALIDATION_ERROR"
  | "UNAUTHORIZED"
  | "FORBIDDEN"
  | "CONFLICT"
  | "INTERNAL_ERROR"
  | "BAD_REQUEST";

export interface ApiError {
  error: {
    code: ErrorCode;
    message: string;
    details?: string[];
  };
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

// ==================== Auth ====================

export interface RegisterRequest {
  email: string;
  full_name: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RefreshRequest {
  refresh_token: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

export interface AuthResponse {
  user: UserResponse;
  access_token: string;
  refresh_token: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
}

// ==================== Users ====================

export interface UserResponse {
  id: string;
  email: string;
  full_name: string;
  is_active: boolean;
  is_email_verified: boolean;
  last_login_at?: string;
  roles: RoleResponse[];
  created_at: string;
  updated_at: string;
}

export interface UpdateUserRequest {
  full_name: string;
  is_active: boolean;
}

export interface AssignRoleRequest {
  role_id: string;
}

export interface RoleResponse {
  id: string;
  name: string;
  display_name: string;
  description: string;
  is_system_role: boolean;
  permissions?: PermissionResponse[];
}

export interface PermissionResponse {
  id: string;
  name: string;
  display_name: string;
  category: string;
}

// ==================== Programs ====================

export type ProgramStatus = "active" | "inactive" | "closed";

export interface CreateProgramRequest {
  name: string;
  description: string;
  status: ProgramStatus;
  start_date?: string;
  end_date?: string;
}

export interface UpdateProgramRequest {
  name: string;
  description: string;
  status: ProgramStatus;
  start_date?: string;
  end_date?: string;
}

export interface ProgramResponse {
  id: string;
  name: string;
  description: string;
  status: ProgramStatus;
  start_date?: string;
  end_date?: string;
  created_at: string;
  updated_at: string;
}

// ==================== Beneficiaries ====================

export type BeneficiaryStatus = "active" | "inactive" | "suspended";

export interface CreateBeneficiaryRequest {
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: string;
  date_of_birth?: string;
  status: BeneficiaryStatus;
  notes?: string;
}

export interface UpdateBeneficiaryRequest {
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: string;
  date_of_birth?: string;
  status: BeneficiaryStatus;
  notes?: string;
}

export interface EnrollProgramRequest {
  program_id: string;
}

export interface BeneficiaryResponse {
  id: string;
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: string;
  date_of_birth?: string;
  status: BeneficiaryStatus;
  notes?: string;
  programs?: ProgramEnrollmentResponse[];
  created_at: string;
  updated_at: string;
}

export interface ProgramEnrollmentResponse {
  program_id: string;
  program_name: string;
  enrolled_at: string;
  status: string;
}

// ==================== Address ====================

export interface AddressDTO {
  street: string;
  city: string;
  state_or_region: string;
  postal_code?: string;
  country?: string;
}

// ==================== Tenants ====================

export interface CreateTenantRequest {
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: AddressDTO;
  notes?: string;
}

export interface UpdateTenantRequest {
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: AddressDTO;
  notes?: string;
}

export interface TenantResponse {
  id: string;
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: AddressDTO;
  landlord_id: string;
  is_active: boolean;
  notes?: string;
  created_at: string;
  updated_at: string;
}

// ==================== Properties ====================

export interface CreatePropertyRequest {
  name: string;
  property_code: string;
  address?: AddressDTO;
  geojson_coordinates?: string;
  property_type?: string;
  size_in_acres?: number;
  size_in_sqm: number;
  monthly_rent_amount?: number;
  description?: string;
}

export interface UpdatePropertyRequest {
  name: string;
  property_code: string;
  address?: AddressDTO;
  geojson_coordinates?: string;
  property_type?: string;
  size_in_acres?: number;
  size_in_sqm: number;
  is_available_for_rent?: boolean;
  monthly_rent_amount?: number;
  description?: string;
}

export interface PropertyResponse {
  id: string;
  name: string;
  property_code: string;
  address?: AddressDTO;
  geojson_coordinates?: string;
  property_type: string;
  size_in_acres?: number;
  size_in_sqm: number;
  owner_id: string;
  is_available_for_rent: boolean;
  is_active: boolean;
  monthly_rent_amount?: number;
  description?: string;
  created_at: string;
  updated_at: string;
}

// ==================== Money ====================

export interface MoneyDTO {
  amount: string;
  currency: string;
}

// ==================== Debts ====================

export type DebtStatus = "PENDING" | "PARTIAL" | "PAID" | "OVERDUE" | "CANCELLED";
export type DebtType = "RENT" | "UTILITIES" | "MAINTENANCE" | "PENALTY" | "OTHER";

export interface CreateDebtRequest {
  tenant_id: string;
  property_id?: string;
  debt_type: DebtType;
  description: string;
  original_amount: MoneyDTO;
  due_date: string;
  notes?: string;
}

export interface UpdateDebtRequest {
  description: string;
  debt_type: DebtType;
  due_date: string;
  property_id?: string;
  notes?: string;
}

export interface CancelDebtRequest {
  reason?: string;
}

export interface DebtResponse {
  id: string;
  tenant_id: string;
  landlord_id: string;
  property_id?: string;
  debt_type: DebtType;
  description: string;
  original_amount: MoneyDTO;
  amount_paid: MoneyDTO;
  balance: MoneyDTO;
  due_date: string;
  status: DebtStatus;
  notes?: string;
  created_at: string;
  updated_at: string;
}

// ==================== Transactions ====================

export interface RecordPaymentRequest {
  debt_id: string;
  tenant_id: string;
  amount: MoneyDTO;
  payment_method: string;
  transaction_date: string;
  description: string;
  receipt_number?: string;
  reference_number?: string;
}

export interface RecordRefundRequest {
  debt_id: string;
  tenant_id: string;
  amount: MoneyDTO;
  payment_method: string;
  refund_date: string;
  description: string;
  reference_number?: string;
}

export interface TransactionResponse {
  id: string;
  debt_id: string;
  tenant_id: string;
  landlord_id: string;
  recorded_by_user_id?: string;
  transaction_type: string;
  amount: MoneyDTO;
  payment_method: string;
  transaction_date: string;
  description: string;
  receipt_number?: string;
  reference_number?: string;
  is_verified: boolean;
  verified_by_user_id?: string;
  verified_at?: string;
  created_at: string;
  updated_at: string;
}

// ==================== Audit ====================

export interface AuditEntryResponse {
  id: string;
  user_id: string;
  user_email: string;
  user_role: string;
  action: string;
  resource_type: string;
  resource_id: string;
  ip_address: string;
  endpoint: string;
  method: string;
  status_code: number;
  success: boolean;
  error_message?: string;
  metadata?: Record<string, unknown>;
  timestamp: string;
}

// ==================== Dashboard ====================

export interface DashboardStatsResponse {
  total_tenants: number;
  total_properties: number;
  active_debts: number;
  overdue_debts: number;
}

export interface RecentActivityResponse {
  action: string;
  description: string;
  timestamp: string;
}

export interface RecentActivitiesResponse {
  data: RecentActivityResponse[];
}

// ==================== Health ====================

export interface HealthResponse {
  status: string;
  timestamp: string;
  services: Record<string, string>;
}
