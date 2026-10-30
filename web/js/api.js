(function () {
  const cfg = window.__APP_CONFIG__ || {
    storagePrefix: "/api/storage",
    searchPrefix: "/api/search",
  };

  function join(base, path) {
    const b = base.replace(/\/$/, "");
    const p = path.startsWith("/") ? path : "/" + path;
    return b + p;
  }

  async function request(base, path, options) {
    const url = join(base, path);
    const res = await fetch(url, {
      headers: { Accept: "application/json", ...(options && options.headers) },
      ...options,
    });
    const text = await res.text();
    let data = null;
    if (text) {
      try {
        data = JSON.parse(text);
      } catch {
        data = { raw: text };
      }
    }
    if (!res.ok) {
      const msg =
        (data && data.error) ||
        res.statusText ||
        "Ошибка запроса";
      const err = new Error(msg);
      err.status = res.status;
      err.body = data;
      throw err;
    }
    return data;
  }

  const storage = {
    listProfiles() {
      return request(cfg.storagePrefix, "/profiles");
    },
    getProfile(id) {
      return request(cfg.storagePrefix, "/profiles/" + encodeURIComponent(id));
    },
    createProfile(name, config) {
      return request(cfg.storagePrefix, "/profiles", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, config }),
      });
    },
    updateProfile(id, name, config) {
      return request(cfg.storagePrefix, "/profiles/" + encodeURIComponent(id), {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, config }),
      });
    },
    deleteProfile(id) {
      return request(cfg.storagePrefix, "/profiles/" + encodeURIComponent(id), {
        method: "DELETE",
      });
    },
    listPurchases(params) {
      const q = new URLSearchParams();
      if (params.law) q.set("law", params.law);
      if (params.q) q.set("q", params.q);
      if (params.offset != null) q.set("offset", String(params.offset));
      if (params.limit != null) q.set("limit", String(params.limit));
      const qs = q.toString();
      return request(cfg.storagePrefix, "/purchases" + (qs ? "?" + qs : ""));
    },
    getPurchase(id) {
      return request(cfg.storagePrefix, "/purchases/" + encodeURIComponent(id));
    },
    runResults(runId, offset, limit) {
      const q = new URLSearchParams();
      if (offset != null) q.set("offset", String(offset));
      if (limit != null) q.set("limit", String(limit));
      const qs = q.toString();
      return request(
        cfg.storagePrefix,
        "/search/runs/" + encodeURIComponent(runId) + "/results" + (qs ? "?" + qs : "")
      );
    },
  };

  const search = {
    createRun(profileId) {
      return request(cfg.searchPrefix, "/runs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: profileId }),
      });
    },
    getRun(id) {
      return request(cfg.searchPrefix, "/runs/" + encodeURIComponent(id));
    },
    runResults(runId, offset, limit) {
      const q = new URLSearchParams();
      if (offset != null) q.set("offset", String(offset));
      if (limit != null) q.set("limit", String(limit));
      const qs = q.toString();
      return request(
        cfg.searchPrefix,
        "/runs/" + encodeURIComponent(runId) + "/results" + (qs ? "?" + qs : "")
      );
    },
    logs(limit) {
      const q = new URLSearchParams();
      if (limit != null) q.set("limit", String(limit));
      const qs = q.toString();
      return request(cfg.searchPrefix, "/logs" + (qs ? "?" + qs : ""));
    },
  };

  window.EisApi = { storage, search, cfg };
})();
