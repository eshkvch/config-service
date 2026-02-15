import "./style.css";
import { createConfigApi } from "./lib/api.js";

const api = createConfigApi();

const state = {
  env: "",
  configs: [],
  selectedKey: null
};

const envForm = document.querySelector("#env-form");
const envInput = document.querySelector("#env-input");
const refreshBtn = document.querySelector("#refresh-btn");
const entryForm = document.querySelector("#entry-form");
const keyInput = document.querySelector("#key-input");
const valueInput = document.querySelector("#value-input");
const createBtn = document.querySelector("#create-btn");
const updateBtn = document.querySelector("#update-btn");
const clearBtn = document.querySelector("#clear-btn");
const lookupForm = document.querySelector("#lookup-form");
const lookupInput = document.querySelector("#lookup-input");
const lookupValue = document.querySelector("#lookup-value");
const lookupMeta = document.querySelector("#lookup-meta");
const statusEl = document.querySelector("#status");
const configList = document.querySelector("#config-list");
const envLabel = document.querySelector("#env-label");
const countLabel = document.querySelector("#count-label");
const editState = document.querySelector("#edit-state");
const clearSelectionBtn = document.querySelector("#clear-selection-btn");
const apiBase = document.querySelector("#api-base");

apiBase.textContent = api.baseUrl;

function setStatus(type, message) {
  statusEl.className = `status status--${type}`;
  statusEl.textContent = message;
}

function setEnv(value) {
  state.env = value;
  envLabel.textContent = value
    ? `Environment: ${value}`
    : "No environment selected";
}

function setEditState(message) {
  editState.textContent = message;
}

function formatUpdatedAt(value) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);
  if (Number.isNaN(date.valueOf())) {
    return value;
  }

  return date.toLocaleString("ru-RU", {
    dateStyle: "short",
    timeStyle: "short"
  });
}

function setLookupResult(config, errorMessage = "") {
  if (!config) {
    lookupValue.textContent = errorMessage || "No lookup yet.";
    lookupMeta.textContent = "";
    return;
  }

  lookupValue.textContent = `${config.key} = ${config.value}`;
  lookupMeta.textContent = `Updated: ${formatUpdatedAt(config.updated_at)}`;
}

function clearEditor() {
  keyInput.value = "";
  valueInput.value = "";
  state.selectedKey = null;
  setEditState("No entry selected.");
}

function renderConfigs() {
  configList.innerHTML = "";
  countLabel.textContent = `${state.configs.length} entries`;

  if (!state.configs.length) {
    const row = document.createElement("tr");
    const cell = document.createElement("td");
    cell.colSpan = 4;
    cell.className = "empty";
    cell.textContent = "No configs in this environment.";
    row.appendChild(cell);
    configList.appendChild(row);
    return;
  }

  state.configs.forEach((config) => {
    const row = document.createElement("tr");

    const keyCell = document.createElement("td");
    keyCell.textContent = config.key;

    const valueCell = document.createElement("td");
    const valueCode = document.createElement("code");
    valueCode.textContent = config.value;
    valueCell.appendChild(valueCode);

    const updatedCell = document.createElement("td");
    updatedCell.textContent = formatUpdatedAt(config.updated_at);

    const actionsCell = document.createElement("td");
    const actions = document.createElement("div");
    actions.className = "actions";

    const selectBtn = document.createElement("button");
    selectBtn.type = "button";
    selectBtn.className = "btn";
    selectBtn.textContent = "Select";
    selectBtn.addEventListener("click", () => {
      keyInput.value = config.key;
      valueInput.value = config.value;
      state.selectedKey = config.key;
      setEditState(`Selected key: ${config.key}`);
    });

    const deleteBtn = document.createElement("button");
    deleteBtn.type = "button";
    deleteBtn.className = "btn btn-danger";
    deleteBtn.textContent = "Delete";
    deleteBtn.addEventListener("click", async () => {
      if (!state.env) {
        setStatus("error", "Set environment before deleting.");
        return;
      }

      if (!window.confirm(`Delete key "${config.key}"?`)) {
        return;
      }

      setStatus("loading", "Deleting...");
      try {
        await api.deleteConfig(state.env, config.key);
        if (state.selectedKey === config.key) {
          clearEditor();
        }
        await loadConfigs({ silent: true });
        setStatus("success", `Deleted ${config.key}.`);
      } catch (error) {
        setStatus("error", error.message || "Delete failed.");
      }
    });

    actions.appendChild(selectBtn);
    actions.appendChild(deleteBtn);
    actionsCell.appendChild(actions);

    row.appendChild(keyCell);
    row.appendChild(valueCell);
    row.appendChild(updatedCell);
    row.appendChild(actionsCell);
    configList.appendChild(row);
  });
}

async function loadConfigs({ silent = false } = {}) {
  if (!state.env) {
    setStatus("error", "Environment is required.");
    return;
  }

  if (!silent) {
    setStatus("loading", "Loading configs...");
  }

  try {
    const configs = await api.listConfigs(state.env);
    state.configs = Array.isArray(configs) ? configs : [];
    renderConfigs();
    if (!silent) {
      setStatus("success", `Loaded ${state.configs.length} entries.`);
    }
  } catch (error) {
    setStatus("error", error.message || "Failed to load configs.");
  }
}

async function checkBackend() {
  setStatus("loading", "Checking API...");
  try {
    await api.health();
    setStatus("success", "API is reachable.");
  } catch (error) {
    setStatus(
      "error",
      "Failed to fetch. Start backend (docker compose up -d --build)."
    );
  }
}

envForm.addEventListener("submit", (event) => {
  event.preventDefault();
  const value = envInput.value.trim();

  if (!value) {
    setStatus("error", "Environment is required.");
    return;
  }

  setEnv(value);
  loadConfigs();
});

refreshBtn.addEventListener("click", () => {
  const value = envInput.value.trim();
  if (value) {
    setEnv(value);
  }
  loadConfigs();
});

createBtn.addEventListener("click", async () => {
  const env = envInput.value.trim();
  const key = keyInput.value.trim();
  const value = valueInput.value;

  if (!env || !key) {
    setStatus("error", "Environment and key are required.");
    return;
  }

  setEnv(env);
  setStatus("loading", "Creating...");

  try {
    await api.createConfig(env, key, value);
    state.selectedKey = key;
    await loadConfigs({ silent: true });
    setStatus("success", `Created ${key}.`);
  } catch (error) {
    setStatus("error", error.message || "Create failed.");
  }
});

updateBtn.addEventListener("click", async () => {
  const env = envInput.value.trim();
  const key = keyInput.value.trim();
  const value = valueInput.value;

  if (!env || !key) {
    setStatus("error", "Environment and key are required.");
    return;
  }

  setEnv(env);
  setStatus("loading", "Updating...");

  try {
    await api.updateConfig(env, key, value);
    state.selectedKey = key;
    await loadConfigs({ silent: true });
    setStatus("success", `Updated ${key}.`);
  } catch (error) {
    setStatus("error", error.message || "Update failed.");
  }
});

clearBtn.addEventListener("click", () => {
  clearEditor();
  setStatus("idle", "");
});

clearSelectionBtn.addEventListener("click", () => {
  clearEditor();
  setStatus("idle", "");
});

lookupForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const env = envInput.value.trim();
  const key = lookupInput.value.trim();

  if (!env || !key) {
    setStatus("error", "Environment and key are required for lookup.");
    return;
  }

  setEnv(env);
  setStatus("loading", "Finding key...");

  try {
    const config = await api.getConfig(env, key);
    setLookupResult(config);
    setStatus("success", `Found ${key}.`);
  } catch (error) {
    setLookupResult(null, error.message || "Not found.");
    setStatus("error", error.message || "Lookup failed.");
  }
});

entryForm.addEventListener("submit", (event) => {
  event.preventDefault();
});

setEnv("");
setEditState("No entry selected.");
setLookupResult(null);
renderConfigs();
checkBackend();
