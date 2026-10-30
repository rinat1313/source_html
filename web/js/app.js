(function () {
  const api = window.EisApi;
  const EIS_MSG = "ЕИС недоступна";
  const PAGE = 10;

  const state = {
    profiles: [],
    selectedId: null,
    editingNew: false,
    runId: null,
    runLimit: 0,
    resultsOffset: 0,
    resultsTotal: 0,
    resultsItems: [],
    pollTimer: null,
    purchasesOffset: 0,
    purchasesTotal: 0,
    purchasesItems: [],
    selectedPurchaseId: null,
  };

  const el = {
    profileList: document.getElementById("profileList"),
    profilesEmpty: document.getElementById("profilesEmpty"),
    profilesError: document.getElementById("profilesError"),
    btnNewProfile: document.getElementById("btnNewProfile"),
    profileForm: document.getElementById("profileForm"),
    settingsEmpty: document.getElementById("settingsEmpty"),
    profileFormError: document.getElementById("profileFormError"),
    btnSaveProfile: document.getElementById("btnSaveProfile"),
    btnDeleteProfile: document.getElementById("btnDeleteProfile"),
    profileLaw: document.getElementById("profileLaw"),
    sectionEis44: document.getElementById("sectionEis44"),
    sectionEis223: document.getElementById("sectionEis223"),
    tabSettings: document.getElementById("tabSettings"),
    tabSearch: document.getElementById("tabSearch"),
    tabPurchases: document.getElementById("tabPurchases"),
    panelSettings: document.getElementById("panelSettings"),
    panelSearch: document.getElementById("panelSearch"),
    panelPurchases: document.getElementById("panelPurchases"),
    searchEmpty: document.getElementById("searchEmpty"),
    searchPanel: document.getElementById("searchPanel"),
    btnRunSearch: document.getElementById("btnRunSearch"),
    runStatusText: document.getElementById("runStatusText"),
    runProgress: document.getElementById("runProgress"),
    eisBanner: document.getElementById("eisBanner"),
    resultsError: document.getElementById("resultsError"),
    resultsEmpty: document.getElementById("resultsEmpty"),
    runResultsList: document.getElementById("runResultsList"),
    btnLoadMoreResults: document.getElementById("btnLoadMoreResults"),
    filterLaw: document.getElementById("filterLaw"),
    filterQuery: document.getElementById("filterQuery"),
    btnRefreshPurchases: document.getElementById("btnRefreshPurchases"),
    purchasesError: document.getElementById("purchasesError"),
    purchasesEmpty: document.getElementById("purchasesEmpty"),
    purchasesList: document.getElementById("purchasesList"),
    btnLoadMorePurchases: document.getElementById("btnLoadMorePurchases"),
    purchaseDetailWrap: document.getElementById("purchaseDetailWrap"),
    purchaseDetailTitle: document.getElementById("purchaseDetailTitle"),
    purchaseDetailBody: document.getElementById("purchaseDetailBody"),
    btnHelp: document.getElementById("btnHelp"),
    helpModal: document.getElementById("helpModal"),
    btnCloseHelp: document.getElementById("btnCloseHelp"),
  };

  function show(elm, visible) {
    if (!elm) return;
    elm.classList.toggle("hidden", !visible);
  }

  function setText(elm, text) {
    if (elm) elm.textContent = text;
  }

  function splitLines(text) {
    return text
      .split(/[\n,;]+/)
      .map((s) => s.trim())
      .filter(Boolean);
  }

  function splitComma(text) {
    return text
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
  }

  function formatDate(iso) {
    if (!iso) return "—";
    try {
      const d = new Date(iso);
      if (Number.isNaN(d.getTime())) return iso;
      return d.toLocaleDateString("ru-RU");
    } catch {
      return iso;
    }
  }

  function parseConfigFromForm() {
    const law = el.profileLaw.value;
    const keywords = splitLines(document.getElementById("profileKeywords").value);
    const limit = parseInt(document.getElementById("profileLimit").value, 10);
    const cfg = { law, keywords, limit };
    if (law === "44" || law === "both") {
      cfg.eis44 = {
        org_region: document.getElementById("eis44OrgRegion").value.trim(),
        document_types: splitComma(document.getElementById("eis44DocTypes").value),
        period: {
          from: document.getElementById("eis44PeriodFrom").value || undefined,
          to: document.getElementById("eis44PeriodTo").value || undefined,
        },
      };
    }
    if (law === "223" || law === "both") {
      cfg.eis223 = {
        entity_types: splitComma(document.getElementById("eis223EntityTypes").value),
        region: document.getElementById("eis223Region").value.trim() || undefined,
        period: {
          from: document.getElementById("eis223PeriodFrom").value || undefined,
          to: document.getElementById("eis223PeriodTo").value || undefined,
        },
      };
    }
    return cfg;
  }

  function fillFormFromConfig(name, config) {
    document.getElementById("profileName").value = name || "";
    const c = config || {};
    el.profileLaw.value = c.law || "44";
    document.getElementById("profileKeywords").value = (c.keywords || []).join("\n");
    document.getElementById("profileLimit").value = c.limit || 50;
    const e44 = c.eis44 || {};
    document.getElementById("eis44OrgRegion").value = e44.org_region || "";
    document.getElementById("eis44DocTypes").value = (e44.document_types || []).join(", ");
    document.getElementById("eis44PeriodFrom").value = (e44.period && e44.period.from) || "";
    document.getElementById("eis44PeriodTo").value = (e44.period && e44.period.to) || "";
    const e223 = c.eis223 || {};
    document.getElementById("eis223EntityTypes").value = (e223.entity_types || []).join(", ");
    document.getElementById("eis223Region").value = e223.region || "";
    document.getElementById("eis223PeriodFrom").value = (e223.period && e223.period.from) || "";
    document.getElementById("eis223PeriodTo").value = (e223.period && e223.period.to) || "";
    updateLawSections();
  }

  function updateLawSections() {
    const law = el.profileLaw.value;
    show(el.sectionEis44, law === "44" || law === "both");
    show(el.sectionEis223, law === "223" || law === "both");
  }

  function parseProfileConfig(raw) {
    if (!raw) return {};
    if (typeof raw === "object") return raw;
    try {
      return JSON.parse(raw);
    } catch {
      return {};
    }
  }

  function renderProfileList() {
    el.profileList.innerHTML = "";
    const hasItems = state.profiles.length > 0;
    show(el.profilesEmpty, !hasItems);
    state.profiles.forEach((p) => {
      const li = document.createElement("li");
      li.className = "profile-item";
      const selectBtn = document.createElement("button");
      selectBtn.type = "button";
      selectBtn.className = "profile-select";
      selectBtn.textContent = p.name;
      selectBtn.setAttribute("aria-current", state.selectedId === p.id ? "true" : "false");
      selectBtn.addEventListener("click", () => selectProfile(p.id));
      const actions = document.createElement("div");
      actions.className = "profile-item-actions";
      const editBtn = document.createElement("button");
      editBtn.type = "button";
      editBtn.className = "btn btn-sm";
      editBtn.textContent = "✎";
      editBtn.title = "Редактировать";
      editBtn.addEventListener("click", () => {
        selectProfile(p.id);
        activateTab("settings");
      });
      actions.appendChild(editBtn);
      li.appendChild(selectBtn);
      li.appendChild(actions);
      el.profileList.appendChild(li);
    });
    updateSearchPanelVisibility();
  }

  async function loadProfiles() {
    show(el.profilesError, false);
    try {
      state.profiles = await api.storage.listProfiles();
      if (!Array.isArray(state.profiles)) state.profiles = [];
      renderProfileList();
    } catch (err) {
      el.profilesError.textContent = err.message || "Не удалось загрузить профили";
      show(el.profilesError, true);
    }
  }

  function selectProfile(id) {
    state.editingNew = false;
    state.selectedId = id;
    const p = state.profiles.find((x) => x.id === id);
    if (!p) return;
    show(el.settingsEmpty, false);
    show(el.profileForm, true);
    fillFormFromConfig(p.name, parseProfileConfig(p.config));
    show(el.profileFormError, false);
    renderProfileList();
    updateSearchPanelVisibility();
  }

  function startNewProfile() {
    state.editingNew = true;
    state.selectedId = null;
    show(el.settingsEmpty, false);
    show(el.profileForm, true);
    fillFormFromConfig("", { law: "44", keywords: [], limit: 50 });
    show(el.profileFormError, false);
    renderProfileList();
    activateTab("settings");
    document.getElementById("profileName").focus();
  }

  async function saveProfile(ev) {
    ev.preventDefault();
    show(el.profileFormError, false);
    const name = document.getElementById("profileName").value.trim();
    if (!name) {
      el.profileFormError.textContent = "Укажите название профиля";
      show(el.profileFormError, true);
      return;
    }
    const config = parseConfigFromForm();
    if (!config.limit || config.limit <= 0) {
      el.profileFormError.textContent = "Лимит должен быть больше нуля";
      show(el.profileFormError, true);
      return;
    }
    try {
      if (state.editingNew || !state.selectedId) {
        const created = await api.storage.createProfile(name, config);
        state.editingNew = false;
        state.selectedId = created.id;
      } else {
        await api.storage.updateProfile(state.selectedId, name, config);
      }
      await loadProfiles();
      if (state.selectedId) selectProfile(state.selectedId);
    } catch (err) {
      el.profileFormError.textContent = err.message || "Ошибка сохранения";
      show(el.profileFormError, true);
    }
  }

  async function deleteProfile() {
    if (!state.selectedId || state.editingNew) return;
    if (!window.confirm("Удалить профиль «" + document.getElementById("profileName").value + "»?")) {
      return;
    }
    try {
      await api.storage.deleteProfile(state.selectedId);
      state.selectedId = null;
      show(el.profileForm, false);
      show(el.settingsEmpty, true);
      await loadProfiles();
      updateSearchPanelVisibility();
    } catch (err) {
      el.profileFormError.textContent = err.message || "Ошибка удаления";
      show(el.profileFormError, true);
    }
  }

  function updateSearchPanelVisibility() {
    const hasProfile = !!state.selectedId && !state.editingNew;
    show(el.searchEmpty, !hasProfile);
    show(el.searchPanel, hasProfile);
    el.btnRunSearch.disabled = !hasProfile;
    el.btnDeleteProfile.disabled = !state.selectedId || state.editingNew;
  }

  function activateTab(name) {
    const map = {
      settings: [el.tabSettings, el.panelSettings],
      search: [el.tabSearch, el.panelSearch],
      purchases: [el.tabPurchases, el.panelPurchases],
    };
    Object.keys(map).forEach((key) => {
      const [tab, panel] = map[key];
      const on = key === name;
      tab.setAttribute("aria-selected", on ? "true" : "false");
      panel.setAttribute("aria-hidden", on ? "false" : "true");
    });
    if (name === "purchases") {
      loadPurchases(true);
    }
  }

  function purchaseNumber(p) {
    if (p.reg_number) return p.reg_number;
    if (p.guid) return p.guid;
    return p.id;
  }

  function renderPurchaseCard(p, options) {
    const li = document.createElement("li");
    const card = document.createElement("article");
    card.className = "purchase-card";
    if (options && options.selectable) {
      card.setAttribute("role", "button");
      card.tabIndex = 0;
      card.setAttribute("aria-selected", state.selectedPurchaseId === p.id ? "true" : "false");
      card.addEventListener("click", () => showPurchaseDetail(p.id));
      card.addEventListener("keydown", (e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          showPurchaseDetail(p.id);
        }
      });
    }
    const h = document.createElement("h4");
    h.textContent = p.title || "Без названия";
    const meta = document.createElement("p");
    meta.className = "purchase-meta";
    meta.textContent =
      "№ " +
      purchaseNumber(p) +
      " · " +
      (p.law === "223" ? "223-ФЗ" : p.law === "44" ? "44-ФЗ" : p.law) +
      " · " +
      formatDate(p.effective_date);
    card.appendChild(h);
    card.appendChild(meta);
    if (p.customer_name) {
      const cust = document.createElement("p");
      cust.className = "purchase-meta";
      cust.textContent = p.customer_name;
      card.appendChild(cust);
    }
    li.appendChild(card);
    return li;
  }

  function parseFileLinks(raw) {
    if (!raw) return [];
    if (Array.isArray(raw)) return raw;
    try {
      const v = typeof raw === "string" ? JSON.parse(raw) : raw;
      return Array.isArray(v) ? v : [];
    } catch {
      return [];
    }
  }

  async function showPurchaseDetail(id) {
    state.selectedPurchaseId = id;
    renderPurchasesList();
    try {
      const p = await api.storage.getPurchase(id);
      el.purchaseDetailTitle.textContent = p.title || "Закупка";
      const parts = [];
      parts.push("<p><strong>Номер:</strong> " + escapeHtml(purchaseNumber(p)) + "</p>");
      parts.push("<p><strong>Закон:</strong> " + escapeHtml(p.law) + "</p>");
      parts.push("<p><strong>Тип:</strong> " + escapeHtml(p.entity_kind || "—") + "</p>");
      parts.push("<p><strong>Заказчик:</strong> " + escapeHtml(p.customer_name || "—") + "</p>");
      if (p.customer_inn) {
        parts.push("<p><strong>ИНН:</strong> " + escapeHtml(p.customer_inn) + "</p>");
      }
      parts.push("<p><strong>Дата:</strong> " + escapeHtml(formatDate(p.effective_date)) + "</p>");
      if (p.description) {
        parts.push("<p><strong>Описание:</strong> " + escapeHtml(p.description) + "</p>");
      }
      const links = parseFileLinks(p.file_links);
      if (links.length) {
        parts.push("<p><strong>Файлы:</strong></p><ul class=\"file-links\">");
        links.forEach((lnk) => {
          const url = lnk.url || lnk.URL || "#";
          const name = lnk.name || lnk.Name || url;
          parts.push(
            '<li><a href="' +
              escapeAttr(url) +
              '" target="_blank" rel="noopener noreferrer">' +
              escapeHtml(name) +
              "</a></li>"
          );
        });
        parts.push("</ul>");
      }
      el.purchaseDetailBody.innerHTML = parts.join("");
      show(el.purchaseDetailWrap, true);
    } catch (err) {
      el.purchaseDetailBody.textContent = err.message || "Ошибка загрузки";
      show(el.purchaseDetailWrap, true);
    }
  }

  function escapeHtml(s) {
    return String(s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  function escapeAttr(s) {
    return escapeHtml(s).replace(/'/g, "&#39;");
  }

  function renderPurchasesList() {
    el.purchasesList.innerHTML = "";
    state.purchasesItems.forEach((p) => {
      el.purchasesList.appendChild(renderPurchaseCard(p, { selectable: true }));
    });
    show(el.purchasesEmpty, state.purchasesItems.length === 0);
    show(
      el.btnLoadMorePurchases,
      state.purchasesItems.length > 0 && state.purchasesItems.length < state.purchasesTotal
    );
  }

  async function loadPurchases(reset) {
    show(el.purchasesError, false);
    if (reset) {
      state.purchasesOffset = 0;
      state.purchasesItems = [];
    }
    try {
      const data = await api.storage.listPurchases({
        law: el.filterLaw.value,
        q: el.filterQuery.value.trim(),
        offset: state.purchasesOffset,
        limit: PAGE,
      });
      const items = data.items || [];
      state.purchasesTotal = data.total != null ? data.total : items.length;
      if (reset) state.purchasesItems = items;
      else state.purchasesItems = state.purchasesItems.concat(items);
      state.purchasesOffset = state.purchasesItems.length;
      renderPurchasesList();
    } catch (err) {
      el.purchasesError.textContent = err.message || "Не удалось загрузить закупки";
      show(el.purchasesError, true);
    }
  }

  function renderRunResults() {
    el.runResultsList.innerHTML = "";
    state.resultsItems.forEach((item) => {
      const p = item.purchase || item;
      el.runResultsList.appendChild(renderPurchaseCard(p, { selectable: false }));
    });
    show(el.resultsEmpty, state.resultsItems.length === 0 && !state.runId);
    show(
      el.btnLoadMoreResults,
      state.resultsItems.length > 0 && state.resultsItems.length < state.resultsTotal
    );
  }

  function updateRunUI(run) {
    if (!run) return;
    const status = run.status || "unknown";
    let statusLabel = status;
    if (status === "pending") statusLabel = "ожидание";
    if (status === "running") statusLabel = "идёт поиск";
    if (status === "completed") statusLabel = "завершён";
    if (status === "failed") statusLabel = "ошибка";
    setText(
      el.runStatusText,
      "Статус: " + statusLabel + (run.error_message ? " — " + run.error_message : "")
    );
    const matched = run.matched_saved != null ? run.matched_saved : 0;
    const limit = run.limit_n != null ? run.limit_n : state.runLimit;
    state.runLimit = limit;
    if (status === "running" || status === "pending") {
      setText(el.runProgress, "Загружено " + matched + " / лимит " + limit + " · идёт поиск");
    } else {
      setText(el.runProgress, "Загружено " + matched + " / лимит " + limit);
    }
    const showBanner =
      status === "failed" &&
      run.error_message &&
      run.error_message.indexOf(EIS_MSG) !== -1;
    show(el.eisBanner, showBanner);
    el.btnRunSearch.disabled = status === "running" || status === "pending";
  }

  async function fetchRunResults(append) {
    if (!state.runId) return;
    show(el.resultsError, false);
    const offset = append ? state.resultsOffset : 0;
    try {
      const data = await api.search.runResults(state.runId, offset, PAGE);
      const items = data.items || [];
      state.resultsTotal = data.total != null ? data.total : items.length;
      if (append) {
        state.resultsItems = state.resultsItems.concat(items);
      } else {
        state.resultsItems = items;
      }
      state.resultsOffset = state.resultsItems.length;
      renderRunResults();
    } catch (err) {
      el.resultsError.textContent = err.message || "Ошибка загрузки результатов";
      show(el.resultsError, true);
    }
  }

  function stopPolling() {
    if (state.pollTimer) {
      clearInterval(state.pollTimer);
      state.pollTimer = null;
    }
  }

  function startPollingRun() {
    stopPolling();
    state.pollTimer = setInterval(async () => {
      if (!state.runId) return;
      try {
        const run = await api.search.getRun(state.runId);
        updateRunUI(run);
        await fetchRunResults(false);
        if (run.status === "completed" || run.status === "failed") {
          stopPolling();
          el.btnRunSearch.disabled = false;
        }
      } catch {
        stopPolling();
        el.btnRunSearch.disabled = false;
      }
    }, 2000);
  }

  async function runSearch() {
    if (!state.selectedId) return;
    stopPolling();
    state.runId = null;
    state.resultsItems = [];
    state.resultsOffset = 0;
    state.resultsTotal = 0;
    renderRunResults();
    show(el.eisBanner, false);
    show(el.resultsError, false);
    setText(el.runStatusText, "Запуск…");
    setText(el.runProgress, "");
    el.btnRunSearch.disabled = true;
    try {
      const run = await api.search.createRun(state.selectedId);
      state.runId = run.id;
      state.runLimit = run.limit_n || 0;
      updateRunUI(run);
      await fetchRunResults(false);
      startPollingRun();
      activateTab("search");
    } catch (err) {
      el.resultsError.textContent = err.message || "Не удалось запустить поиск";
      show(el.resultsError, true);
      el.btnRunSearch.disabled = false;
    }
  }

  function openHelp() {
    el.helpModal.hidden = false;
    el.btnCloseHelp.focus();
  }

  function closeHelp() {
    el.helpModal.hidden = true;
    el.btnHelp.focus();
  }

  el.btnNewProfile.addEventListener("click", startNewProfile);
  el.profileForm.addEventListener("submit", saveProfile);
  el.btnDeleteProfile.addEventListener("click", deleteProfile);
  el.profileLaw.addEventListener("change", updateLawSections);
  el.tabSettings.addEventListener("click", () => activateTab("settings"));
  el.tabSearch.addEventListener("click", () => activateTab("search"));
  el.tabPurchases.addEventListener("click", () => activateTab("purchases"));
  el.btnRunSearch.addEventListener("click", runSearch);
  el.btnLoadMoreResults.addEventListener("click", () => fetchRunResults(true));
  el.btnRefreshPurchases.addEventListener("click", () => loadPurchases(true));
  el.btnLoadMorePurchases.addEventListener("click", () => loadPurchases(false));
  el.btnHelp.addEventListener("click", openHelp);
  el.btnCloseHelp.addEventListener("click", closeHelp);
  el.helpModal.addEventListener("click", (e) => {
    if (e.target === el.helpModal) closeHelp();
  });
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && !el.helpModal.hidden) closeHelp();
  });

  updateLawSections();
  loadProfiles();
  activateTab("settings");
})();
