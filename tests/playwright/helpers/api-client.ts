const API_BASE = process.env.API_URL || "http://localhost:8080/api/v1";

interface AuthTokens {
  access_token: string;
  refresh_token: string;
}

interface AuthResponse {
  user: { id: string; email: string; full_name: string; roles: { id: string; name: string }[] };
  access_token: string;
  refresh_token: string;
}

async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API ${options.method || "GET"} ${path} failed (${res.status}): ${body}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export class TestApiClient {
  private tokens: AuthTokens | null = null;

  private authHeaders(): Record<string, string> {
    if (!this.tokens) return {};
    return { Authorization: `Bearer ${this.tokens.access_token}` };
  }

  async register(email: string, fullName: string, password: string): Promise<AuthResponse> {
    const res = await apiRequest<AuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, full_name: fullName, password }),
    });
    this.tokens = { access_token: res.access_token, refresh_token: res.refresh_token };
    return res;
  }

  async login(email: string, password: string): Promise<AuthResponse> {
    const res = await apiRequest<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    this.tokens = { access_token: res.access_token, refresh_token: res.refresh_token };
    return res;
  }

  async createTenant(data: { full_name: string; email?: string; phone_number?: string }): Promise<{ id: string }> {
    return apiRequest("/tenants", {
      method: "POST",
      body: JSON.stringify(data),
      headers: this.authHeaders(),
    });
  }

  async createProperty(data: {
    name: string;
    property_code: string;
    property_type: string;
    size_in_sqm: number;
  }): Promise<{ id: string }> {
    return apiRequest("/properties", {
      method: "POST",
      body: JSON.stringify(data),
      headers: this.authHeaders(),
    });
  }

  async createDebt(data: {
    tenant_id: string;
    debt_type: string;
    description: string;
    original_amount: { amount: string; currency: string };
    due_date: string;
  }): Promise<{ id: string }> {
    return apiRequest("/debts", {
      method: "POST",
      body: JSON.stringify(data),
      headers: this.authHeaders(),
    });
  }

  async payDebt(data: {
    debt_id: string;
    tenant_id: string;
    amount: { amount: string; currency: string };
    payment_method: string;
    transaction_date: string;
    description: string;
  }): Promise<{ id: string }> {
    return apiRequest("/transactions/payment", {
      method: "POST",
      body: JSON.stringify(data),
      headers: this.authHeaders(),
    });
  }

  getTokens(): AuthTokens | null {
    return this.tokens;
  }
}
