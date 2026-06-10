import { afterEach, describe, expect, it, vi } from "vitest";
import { createConfigApi, normalizeBaseUrl, pushMetric } from "../src/lib/api.js";

async function flushAsyncHandlers() {
  await new Promise((resolve) => {
    setTimeout(resolve, 0);
  });
}

describe("normalizeBaseUrl additional cases", () => {
  it("uses same-origin for empty values", () => {
    expect(normalizeBaseUrl()).toBe("");
    expect(normalizeBaseUrl("")).toBe("");
  });

  it("keeps URLs without trailing slashes", () => {
    expect(normalizeBaseUrl("https://api.example.test/base"))
      .toBe("https://api.example.test/base");
  });
});

describe("createConfigApi additional behavior", () => {
  it("exposes normalized base URL", () => {
    const api = createConfigApi({ baseUrl: "/" });
    expect(api.baseUrl).toBe("same-origin (/)");
  });

  it("checks health", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ status: "ok" }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      })
    );
    const api = createConfigApi({ fetcher });

    await expect(api.health()).resolves.toEqual({ status: "ok" });
    expect(fetcher).toHaveBeenCalledWith("/health", {
      method: "GET",
      headers: {},
      body: undefined
    });
  });

  it("encodes environment and key path segments", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ key: "a/b", value: "x" }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      })
    );
    const api = createConfigApi({ fetcher });

    await api.getConfig("prod env", "a/b");

    expect(fetcher.mock.calls[0][0])
      .toBe("/api/configs/prod%20env/a%2Fb");
  });

  it("returns null for no-content responses", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(undefined, { status: 204 })
    );
    const api = createConfigApi({ fetcher });

    await expect(api.deleteConfig("prod", "key")).resolves.toBeNull();
  });

  it("uses status fallback for empty error responses", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response("", { status: 503 })
    );
    const api = createConfigApi({ fetcher });

    await expect(api.health()).rejects.toMatchObject({
      message: "Request failed with status 503",
      status: 503
    });
  });

  it("wraps network errors with base URL details", async () => {
    const fetcher = vi.fn().mockRejectedValue(new Error("connection refused"));
    const api = createConfigApi({ baseUrl: "http://api/", fetcher });

    await expect(api.health()).rejects.toThrow(
      "Failed to fetch API (http://api). connection refused"
    );
  });

  it("wraps non-Error network failures", async () => {
    const fetcher = vi.fn().mockRejectedValue("offline");
    const api = createConfigApi({ fetcher });

    await expect(api.health()).rejects.toThrow(
      "Failed to fetch API (same-origin (/)). unknown network error"
    );
  });
});

describe("pushMetric", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("posts prometheus text format", () => {
    const fetch = vi.fn().mockResolvedValue(new Response("", { status: 202 }));
    vi.stubGlobal("fetch", fetch);

    pushMetric("frontend_page_loads_total", 1, { route: "home" });

    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:9091/metrics/job/frontend/instance/1",
      {
        method: "POST",
        body: "# TYPE frontend_page_loads_total gauge\nfrontend_page_loads_total{route=\"home\"} 1\n"
      }
    );
  });

  it("logs metric push failures", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("down")));
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

    pushMetric("frontend_errors_total", 1);
    await flushAsyncHandlers();

    expect(consoleError).toHaveBeenCalledWith(
      "Failed to push metric:",
      expect.any(Error)
    );
  });
});
