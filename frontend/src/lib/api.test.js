import { describe, expect, it, vi } from "vitest";
import { createConfigApi, normalizeBaseUrl } from "./api.js";

describe("normalizeBaseUrl", () => {
  it("removes trailing slashes", () => {
    expect(normalizeBaseUrl("http://localhost:8080/"))
      .toBe("http://localhost:8080");
    expect(normalizeBaseUrl("http://localhost:8080///"))
      .toBe("http://localhost:8080");
  });

  it("uses same-origin when base is slash", () => {
    expect(normalizeBaseUrl("/")).toBe("");
  });
});

describe("createConfigApi", () => {
  it("lists configs with GET", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(JSON.stringify([{ key: "a", value: "b" }]), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      })
    );

    const api = createConfigApi({
      baseUrl: "http://localhost:8080/",
      fetcher
    });

    const result = await api.listConfigs("prod");
    const [url, options] = fetcher.mock.calls[0];

    expect(url).toBe("http://localhost:8080/configs/prod");
    expect(options.method).toBe("GET");
    expect(result).toEqual([{ key: "a", value: "b" }]);
  });

  it("creates configs with POST", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response("", { status: 201 })
    );

    const api = createConfigApi({ baseUrl: "http://api", fetcher });

    await api.createConfig("stage", "token", "secret");
    const [url, options] = fetcher.mock.calls[0];

    expect(url).toBe("http://api/configs/stage/token");
    expect(options.method).toBe("POST");
    expect(JSON.parse(options.body)).toEqual({ value: "secret" });
  });

  it("updates configs with PUT", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(null, { status: 204 })
    );

    const api = createConfigApi({ baseUrl: "http://api", fetcher });

    await api.updateConfig("stage", "token", "new-secret");
    const [url, options] = fetcher.mock.calls[0];

    expect(url).toBe("http://api/configs/stage/token");
    expect(options.method).toBe("PUT");
    expect(JSON.parse(options.body)).toEqual({ value: "new-secret" });
  });

  it("deletes configs with DELETE", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(null, { status: 204 })
    );

    const api = createConfigApi({ baseUrl: "http://api", fetcher });

    await api.deleteConfig("stage", "token");
    const [url, options] = fetcher.mock.calls[0];

    expect(url).toBe("http://api/configs/stage/token");
    expect(options.method).toBe("DELETE");
  });

  it("throws when the server responds with an error", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response("boom", { status: 500 })
    );

    const api = createConfigApi({ baseUrl: "http://api", fetcher });

    await expect(api.getConfig("prod", "missing")).rejects.toThrow("boom");
  });
});

