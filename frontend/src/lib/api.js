const DEFAULT_BASE_URL =
  import.meta.env?.VITE_API_BASE_URL || "/";

export function normalizeBaseUrl(baseUrl) {
  if (!baseUrl || baseUrl === "/") {
    return "";
  }
  return baseUrl.replace(/\/+$/, "");
}

function encodeSegment(value) {
  return encodeURIComponent(value);
}

async function readErrorMessage(response) {
  const text = await response.text();
  return text || `Request failed with status ${response.status}`;
}

export function createConfigApi({ baseUrl = DEFAULT_BASE_URL, fetcher = fetch } = {}) {
  const root = normalizeBaseUrl(baseUrl);

  async function request(path, options = {}) {
    const { method = "GET", body, parse = "json" } = options;
    const headers = {};
    const hasBody = body !== undefined;

    if (hasBody) {
      headers["Content-Type"] = "application/json";
    }

    let response;
    try {
      response = await fetcher(`${root}${path}`, {
        method,
        headers,
        body: hasBody ? JSON.stringify(body) : undefined
      });
    } catch (error) {
      const detail =
        error instanceof Error ? error.message : "unknown network error";
      throw new Error(
        `Failed to fetch API (${root || "same-origin (/)"}). ${detail}`
      );
    }

    if (!response.ok) {
      const message = await readErrorMessage(response);
      const error = new Error(message);
      error.status = response.status;
      throw error;
    }

    if (parse === "none" || response.status === 204) {
      return null;
    }

    return response.json();
  }

  return {
    baseUrl: root || "same-origin (/)",
    health: () => request("/health"),
    listConfigs: (env) => request(`/configs/${encodeSegment(env)}`),
    getConfig: (env, key) =>
      request(`/configs/${encodeSegment(env)}/${encodeSegment(key)}`),
    createConfig: (env, key, value) =>
      request(`/configs/${encodeSegment(env)}/${encodeSegment(key)}`, {
        method: "POST",
        body: { value },
        parse: "none"
      }),
    updateConfig: (env, key, value) =>
      request(`/configs/${encodeSegment(env)}/${encodeSegment(key)}`, {
        method: "PUT",
        body: { value },
        parse: "none"
      }),
    deleteConfig: (env, key) =>
      request(`/configs/${encodeSegment(env)}/${encodeSegment(key)}`, {
        method: "DELETE",
        parse: "none"
      })
  };
}


