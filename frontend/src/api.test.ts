import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, api } from "./api";

describe("API client", () => {
  afterEach(() => vi.restoreAllMocks());

  it("sends a create request using the documented internal route", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ name: "new-checkout", description: "", status: "open", statusDate: "2026-01-01T00:00:00Z", whitelist: [] }), { status: 201 }));
    await api.createFeature({ name: "new-checkout", description: "", status: "open", whitelist: [] });
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/internal/features", expect.objectContaining({ method: "POST" }));
  });

  it("turns dependency failures into a typed error", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ error: "dependency unavailable" }), { status: 503 }));
    await expect(api.listFeatures()).rejects.toEqual(expect.objectContaining<ApiError>({ name: "ApiError", status: 503, message: "dependency unavailable" }));
  });
});
