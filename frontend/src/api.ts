export type FeatureStatus = "open" | "closed" | "whitelisted";

export interface Feature {
  name: string;
  description: string;
  status: FeatureStatus;
  statusDate: string;
  whitelist: string[];
}

export interface FeatureInput {
  name?: string;
  description: string;
  status: FeatureStatus;
  whitelist: string[];
}

export interface Evaluation {
  featureName: string;
  userId: string;
  enabled: boolean;
  status: FeatureStatus;
}

export class ApiError extends Error {
  constructor(public readonly status: number, message: string) {
    super(message);
    this.name = "ApiError";
  }
}

const baseUrl = (import.meta.env.VITE_API_BASE_URL ?? "").replace(/\/$/, "");

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${baseUrl}${path}`, {
      ...init,
      headers: { "Content-Type": "application/json", ...init?.headers }
    });
  } catch {
    throw new ApiError(0, "The API could not be reached. Check that the service is running.");
  }

  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as { error?: string } | null;
    throw new ApiError(response.status, body?.error ?? `Request failed (${response.status}).`);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export const api = {
  listFeatures: () => request<Feature[]>("/api/v1/internal/features"),
  getFeature: (name: string) => request<Feature>(`/api/v1/internal/features/${encodeURIComponent(name)}`),
  createFeature: (input: FeatureInput) => request<Feature>("/api/v1/internal/features", { method: "POST", body: JSON.stringify(input) }),
  updateFeature: (name: string, input: FeatureInput) => request<Feature>(`/api/v1/internal/features/${encodeURIComponent(name)}`, { method: "PUT", body: JSON.stringify(input) }),
  deleteFeature: (name: string) => request<void>(`/api/v1/internal/features/${encodeURIComponent(name)}`, { method: "DELETE" }),
  refreshCache: () => request<{ refreshed: number }>("/api/v1/internal/cache/refresh", { method: "POST" }),
  evaluate: (name: string, userId: string) => request<Evaluation>(`/api/v1/external/features/${encodeURIComponent(name)}/users/${encodeURIComponent(userId)}/evaluation`)
};
