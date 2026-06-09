import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

class FakeElement {
  constructor(tagName = "div") {
    this.tagName = tagName.toUpperCase();
    this.children = [];
    this.listeners = new Map();
    this.className = "";
    this.textContent = "";
    this.value = "";
    this.type = "";
    this.colSpan = 0;
    this._innerHTML = "";
  }

  set innerHTML(value) {
    this._innerHTML = value;
    if (value === "") {
      this.children = [];
    }
  }

  get innerHTML() {
    return this._innerHTML;
  }

  appendChild(child) {
    this.children.push(child);
    return child;
  }

  addEventListener(type, handler) {
    const handlers = this.listeners.get(type) || [];
    handlers.push(handler);
    this.listeners.set(type, handlers);
  }

  dispatch(type) {
    const event = {
      type,
      defaultPrevented: false,
      preventDefault() {
        this.defaultPrevented = true;
      }
    };

    for (const handler of this.listeners.get(type) || []) {
      handler(event);
    }
    return event;
  }

  click() {
    return this.dispatch("click");
  }
}

class FakeDocument {
  constructor() {
    this.created = [];
    this.elements = new Map(
      [
        "env-form",
        "env-input",
        "refresh-btn",
        "entry-form",
        "key-input",
        "value-input",
        "create-btn",
        "update-btn",
        "clear-btn",
        "lookup-form",
        "lookup-input",
        "lookup-value",
        "lookup-meta",
        "status",
        "config-list",
        "env-label",
        "count-label",
        "edit-state",
        "clear-selection-btn",
        "api-base"
      ].map((id) => [id, new FakeElement()])
    );
  }

  querySelector(selector) {
    if (!selector.startsWith("#")) {
      throw new Error(`Unsupported selector: ${selector}`);
    }

    const element = this.elements.get(selector.slice(1));
    if (!element) {
      throw new Error(`Missing fake element: ${selector}`);
    }
    return element;
  }

  createElement(tagName) {
    const element = new FakeElement(tagName);
    this.created.push(element);
    return element;
  }

  byId(id) {
    return this.elements.get(id);
  }
}

function jsonResponse(value, init = {}) {
  return new Response(JSON.stringify(value), {
    status: init.status || 200,
    headers: { "Content-Type": "application/json" }
  });
}

async function flushPromises() {
  for (let i = 0; i < 10; i += 1) {
    await Promise.resolve();
  }
  await new Promise((resolve) => {
    setTimeout(resolve, 0);
  });
}

async function importMainWithDom({ fetcher, confirm = true } = {}) {
  const document = new FakeDocument();
  const windowListeners = new Map();
  const window = {
    confirm: vi.fn(() => confirm),
    addEventListener: vi.fn((type, handler) => {
      windowListeners.set(type, handler);
    })
  };

  vi.stubGlobal("document", document);
  vi.stubGlobal("window", window);
  vi.stubGlobal("fetch", fetcher);
  vi.resetModules();

  await import("../src/main.js");
  await flushPromises();

  return { document, window, windowListeners };
}

describe("main UI flow", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("loads, renders, selects, creates, updates, looks up, deletes, and clears configs", async () => {
    const configs = [
      {
        key: "token",
        value: "secret",
        updated_at: "2026-06-09T10:00:00Z"
      }
    ];
    const fetcher = vi.fn(async (url, options = {}) => {
      const method = options.method || "GET";
      if (url === "/health") {
        return jsonResponse({ status: "ok" });
      }
      if (String(url).startsWith("http://localhost:9091")) {
        return new Response("", { status: 202 });
      }
      if (url === "/api/configs/prod" && method === "GET") {
        return jsonResponse(configs);
      }
      if (url === "/api/configs/prod/token" && method === "GET") {
        return jsonResponse(configs[0]);
      }
      if (url === "/api/configs/prod/token" && method === "POST") {
        return new Response("", { status: 201 });
      }
      if (url === "/api/configs/prod/token" && method === "PUT") {
        return new Response(undefined, { status: 204 });
      }
      if (url === "/api/configs/prod/token" && method === "DELETE") {
        return new Response(undefined, { status: 204 });
      }
      throw new Error(`Unexpected request ${method} ${url}`);
    });

    const { document, window, windowListeners } = await importMainWithDom({ fetcher });

    expect(document.byId("api-base").textContent).toBe("same-origin (/)");
    expect(document.byId("status").textContent).toBe("API is reachable.");
    expect(windowListeners.has("error")).toBe(true);
    windowListeners.get("error")();
    expect(fetcher).toHaveBeenCalledWith(
      "http://localhost:9091/metrics/job/frontend/instance/1",
      expect.objectContaining({
        body: "# TYPE frontend_errors_total gauge\nfrontend_errors_total{} 1\n"
      })
    );

    document.byId("env-input").value = "prod";
    document.byId("env-form").dispatch("submit");
    await flushPromises();

    expect(document.byId("env-label").textContent).toBe("Environment: prod");
    expect(document.byId("count-label").textContent).toBe("1 entries");
    expect(document.byId("status").textContent).toBe("Loaded 1 entries.");

    const row = document.byId("config-list").children[0];
    expect(row.children[0].textContent).toBe("token");
    expect(row.children[1].children[0].textContent).toBe("secret");
    expect(row.children[2].textContent).not.toBe("-");

    const selectButton = row.children[3].children[0].children[0];
    selectButton.click();
    expect(document.byId("key-input").value).toBe("token");
    expect(document.byId("value-input").value).toBe("secret");
    expect(document.byId("edit-state").textContent).toBe("Selected key: token");

    document.byId("create-btn").click();
    await flushPromises();
    expect(fetcher).toHaveBeenCalledWith(
      "/api/configs/prod/token",
      expect.objectContaining({ method: "POST" })
    );
    expect(document.byId("status").textContent).toBe("Created token.");

    document.byId("update-btn").click();
    await flushPromises();
    expect(fetcher).toHaveBeenCalledWith(
      "/api/configs/prod/token",
      expect.objectContaining({ method: "PUT" })
    );
    expect(document.byId("status").textContent).toBe("Updated token.");

    document.byId("lookup-input").value = "token";
    document.byId("lookup-form").dispatch("submit");
    await flushPromises();
    expect(document.byId("lookup-value").textContent).toBe("token = secret");
    expect(document.byId("lookup-meta").textContent).toContain("Updated:");
    expect(document.byId("status").textContent).toBe("Found token.");

    const deleteButton = document.byId("config-list").children[0].children[3].children[0].children[1];
    deleteButton.click();
    await flushPromises();
    expect(window.confirm).toHaveBeenCalledWith("Delete key \"token\"?");
    expect(fetcher).toHaveBeenCalledWith(
      "/api/configs/prod/token",
      expect.objectContaining({ method: "DELETE" })
    );
    expect(document.byId("edit-state").textContent).toBe("No entry selected.");
    expect(document.byId("status").textContent).toBe("Deleted token.");

    document.byId("clear-btn").click();
    expect(document.byId("key-input").value).toBe("");
    expect(document.byId("value-input").value).toBe("");
    expect(document.byId("status").textContent).toBe("");

    document.byId("clear-selection-btn").click();
    expect(document.byId("edit-state").textContent).toBe("No entry selected.");
  });

  it("shows validation and API errors", async () => {
    const fetcher = vi.fn(async (url, options = {}) => {
      const method = options.method || "GET";
      if (url === "/health") {
        throw new Error("offline");
      }
      if (String(url).startsWith("http://localhost:9091")) {
        return new Response("", { status: 202 });
      }
      if (url === "/api/configs/prod" && method === "GET") {
        return new Response("list failed", { status: 500 });
      }
      if (url === "/api/configs/prod/token" && method === "GET") {
        return new Response("missing", { status: 404 });
      }
      if (url === "/api/configs/prod/token" && method === "POST") {
        return new Response("create failed", { status: 500 });
      }
      if (url === "/api/configs/prod/token" && method === "PUT") {
        return new Response("update failed", { status: 500 });
      }
      if (url === "/api/configs/prod/token" && method === "DELETE") {
        return new Response("delete failed", { status: 500 });
      }
      throw new Error(`Unexpected request ${method} ${url}`);
    });

    const { document } = await importMainWithDom({ fetcher, confirm: false });

    expect(document.byId("status").textContent)
      .toBe("Failed to fetch. Start backend (docker compose up -d --build).");

    document.byId("env-form").dispatch("submit");
    expect(document.byId("status").textContent).toBe("Environment is required.");

    document.byId("refresh-btn").click();
    expect(document.byId("status").textContent).toBe("Environment is required.");

    document.byId("create-btn").click();
    expect(document.byId("status").textContent).toBe("Environment and key are required.");

    document.byId("update-btn").click();
    expect(document.byId("status").textContent).toBe("Environment and key are required.");

    document.byId("lookup-form").dispatch("submit");
    expect(document.byId("status").textContent)
      .toBe("Environment and key are required for lookup.");

    document.byId("env-input").value = "prod";
    document.byId("key-input").value = "token";
    document.byId("value-input").value = "secret";

    document.byId("env-form").dispatch("submit");
    await flushPromises();
    expect(document.byId("status").textContent).toBe("list failed");

    document.byId("create-btn").click();
    await flushPromises();
    expect(document.byId("status").textContent).toBe("create failed");

    document.byId("update-btn").click();
    await flushPromises();
    expect(document.byId("status").textContent).toBe("update failed");

    document.byId("lookup-input").value = "token";
    document.byId("lookup-form").dispatch("submit");
    await flushPromises();
    expect(document.byId("lookup-value").textContent).toBe("missing");
    expect(document.byId("status").textContent).toBe("missing");
  });

  it("handles empty lists, unchanged refreshes, delete cancellation, and delete failures", async () => {
    const configs = [
      {
        key: "bad-date",
        value: "value",
        updated_at: "not-a-date"
      }
    ];
    const fetcher = vi.fn(async (url, options = {}) => {
      const method = options.method || "GET";
      if (url === "/health") {
        return jsonResponse({ status: "ok" });
      }
      if (String(url).startsWith("http://localhost:9091")) {
        return new Response("", { status: 202 });
      }
      if (url === "/api/configs/prod" && method === "GET") {
        return jsonResponse(configs);
      }
      if (url === "/api/configs/prod/bad-date" && method === "DELETE") {
        return new Response("delete failed", { status: 500 });
      }
      throw new Error(`Unexpected request ${method} ${url}`);
    });

    const { document, window } = await importMainWithDom({ fetcher, confirm: false });

    document.byId("env-input").value = "prod";
    document.byId("env-form").dispatch("submit");
    await flushPromises();

    expect(document.byId("config-list").children[0].children[2].textContent)
      .toBe("not-a-date");

    document.byId("env-input").value = "";
    document.byId("refresh-btn").click();
    await flushPromises();
    expect(document.byId("status").textContent).toBe("Loaded 1 entries.");

    const deleteButton = document.byId("config-list").children[0].children[3].children[0].children[1];
    deleteButton.click();
    await flushPromises();
    expect(window.confirm).toHaveBeenCalledWith("Delete key \"bad-date\"?");
    expect(fetcher).not.toHaveBeenCalledWith(
      "/api/configs/prod/bad-date",
      expect.objectContaining({ method: "DELETE" })
    );

    window.confirm.mockReturnValue(true);
    deleteButton.click();
    await flushPromises();
    expect(document.byId("status").textContent).toBe("delete failed");
  });
});
