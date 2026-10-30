(function () {
  var statusEl = document.getElementById("config-status");
  fetch("/config.json")
    .then(function (r) {
      if (!r.ok) {
        throw new Error("HTTP " + r.status);
      }
      return r.json();
    })
    .then(function (cfg) {
      window.EIS_UI_CONFIG = cfg;
      statusEl.className = "ok";
      statusEl.textContent =
        "API storage: " +
        cfg.storageAPIBase +
        " · API search: " +
        cfg.searchAPIBase;
    })
    .catch(function (e) {
      statusEl.className = "err";
      statusEl.textContent = "Не удалось загрузить config.json: " + e.message;
    });
})();
