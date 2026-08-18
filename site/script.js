const rawCatalog = (window.FRESHCTL_CATALOG || []).map((pkg) => ({ ...pkg, id: pkg.packageId }));
const packageCatalog = [...rawCatalog].sort((left, right) =>
  left.name.localeCompare(right.name, undefined, { sensitivity: "base" }),
);
const translations = window.FRESHCTL_I18N || {};
const russianCatalog = window.FRESHCTL_RU_CATALOG || { categories: {}, types: {}, descriptions: {} };

function getInitialLanguage() {
  try {
    const saved = window.localStorage.getItem("freshctl-language");
    if (saved === "en" || saved === "ru") return saved;
  } catch {
    // The browser may block local storage in private or restricted contexts.
  }
  return navigator.language?.toLowerCase().startsWith("ru") ? "ru" : "en";
}

let currentLanguage = getInitialLanguage();

function translate(key, values = {}) {
  const template = translations[currentLanguage]?.[key] ?? translations.en?.[key] ?? key;
  return Object.entries(values).reduce(
    (result, [name, value]) => result.replaceAll(`{${name}}`, String(value)),
    template,
  );
}

function localizedDescription(pkg) {
  if (currentLanguage === "ru") return russianCatalog.descriptions?.[pkg.packageId] || pkg.description;
  return pkg.description;
}

function localizedCategory(pkg) {
  if (currentLanguage === "ru") return russianCatalog.categories?.[pkg.category] || pkg.category;
  return pkg.category;
}

function localizedType(pkg) {
  if (currentLanguage === "ru") return russianCatalog.types?.[pkg.type] || pkg.type;
  return pkg.type;
}

const copyButtons = document.querySelectorAll("[data-copy]");
const packageModal = document.querySelector("#package-modal");
const packageSearch = document.querySelector("#package-search");
const packageList = document.querySelector("#package-list");
const packageCount = document.querySelector("#package-count");
const packageOpeners = document.querySelectorAll("[data-open-packages]");
const packageClosers = document.querySelectorAll("[data-close-packages]");
const installSpotlightLink = document.querySelector("[data-spotlight-install]");
const installSpotlight = document.querySelector("[data-close-install-spotlight]");
const installCard = document.querySelector("[data-install-card]");
const profileSearch = document.querySelector("#profile-search");
const profileCount = document.querySelector("#profile-count");
const profilePackageList = document.querySelector("#profile-package-list");
const profileOutput = document.querySelector("#profile-json-output");
const profileHint = document.querySelector("#profile-hint");
const profileCopy = document.querySelector("#profile-copy");
const profileDownload = document.querySelector("#profile-download");
const languageButtons = document.querySelectorAll("[data-language]");
let installSpotlightTimer;
const selectedProfilePackages = new Set();

function applyLanguage(language, persist = false) {
  currentLanguage = language === "ru" ? "ru" : "en";
  document.documentElement.lang = currentLanguage;
  document.title = translate("meta.title");
  document.querySelector('meta[name="description"]')?.setAttribute("content", translate("meta.description"));

  document.querySelectorAll("[data-i18n]").forEach((element) => {
    element.textContent = translate(element.dataset.i18n);
  });
  document.querySelectorAll("[data-i18n-placeholder]").forEach((element) => {
    element.setAttribute("placeholder", translate(element.dataset.i18nPlaceholder));
  });
  document.querySelectorAll("[data-i18n-aria-label]").forEach((element) => {
    element.setAttribute("aria-label", translate(element.dataset.i18nAriaLabel));
  });
  languageButtons.forEach((button) => {
    button.setAttribute("aria-pressed", String(button.dataset.language === currentLanguage));
  });

  if (persist) {
    try {
      window.localStorage.setItem("freshctl-language", currentLanguage);
    } catch {
      // Language switching still works when local storage is unavailable.
    }
  }

  renderPackages();
  renderProfileGenerator();
}

copyButtons.forEach((button) => {
  button.addEventListener("click", async () => {
    const text = button.getAttribute("data-copy");
    if (!text) return;

    try {
      await navigator.clipboard.writeText(text);
      button.textContent = translate("common.copied");
      button.classList.add("copied");

      window.setTimeout(() => {
        button.textContent = translate(button.dataset.i18n || "common.copy");
        button.classList.remove("copied");
      }, 1400);
    } catch {
      button.textContent = translate("common.copyFailed");
      window.setTimeout(() => {
        button.textContent = translate(button.dataset.i18n || "common.copy");
      }, 1400);
    }
  });
});

function renderPackages() {
  if (!packageList || !packageCount || !packageSearch) return;

  const query = packageSearch.value.trim().toLowerCase();
  const visible = packageCatalog.filter((pkg) => {
    const haystack = `${pkg.name} ${pkg.packageId} ${localizedDescription(pkg)} ${localizedCategory(pkg)} ${localizedType(pkg)}`.toLowerCase();
    return haystack.includes(query);
  });

  packageCount.textContent = translate("catalog.count", { shown: visible.length, total: packageCatalog.length });

  if (visible.length === 0) {
    packageList.innerHTML = `<div class="package-empty">${escapeHtml(translate("catalog.empty"))}</div>`;
    return;
  }

  packageList.innerHTML = visible
    .map(
      (pkg, index) => `
        <div class="package-row" style="animation-delay: ${Math.min(index, 10) * 16}ms">
          <strong>${escapeHtml(pkg.name)}</strong>
          <code>${escapeHtml(pkg.packageId)}</code>
        </div>
      `,
    )
    .join("");
}

function renderProfileGenerator() {
  if (!profilePackageList || !profileCount || !profileSearch || !profileOutput) return;

  const query = profileSearch.value.trim().toLowerCase();
  const visible = rawCatalog.filter((pkg) => matchesProfileQuery(pkg, query));
  const selectedCount = selectedProfilePackages.size;
  const profile = buildProfile();
  const hasSelection = profile.packages.length > 0;

  profileCount.textContent = translate("profile.count", { selected: selectedCount, shown: visible.length });
  profileOutput.textContent = hasSelection ? JSON.stringify(profile, null, 2) : "";
  if (profileHint) {
    profileHint.textContent = hasSelection
      ? translate("profile.readyHint")
      : translate("profile.emptyHint");
  }
  if (profileCopy) profileCopy.disabled = !hasSelection;
  if (profileDownload) profileDownload.disabled = !hasSelection;

  if (visible.length === 0) {
    profilePackageList.innerHTML = `<div class="package-empty">${escapeHtml(translate("catalog.empty"))}</div>`;
    return;
  }

  profilePackageList.innerHTML = visible
    .map((pkg) => {
      const selected = selectedProfilePackages.has(pkg.packageId);
      return `
        <button class="profile-package-row${selected ? " selected" : ""}" type="button" data-profile-package="${escapeHtml(pkg.packageId)}">
          <span>
            <strong>${escapeHtml(pkg.name)}</strong>
            <small>${escapeHtml(localizedDescription(pkg) || pkg.packageId)}</small>
          </span>
          <code>${selected ? escapeHtml(translate("profile.selected")) : pkg.packageId}</code>
        </button>
      `;
    })
    .join("");
}

function matchesProfileQuery(pkg, query) {
  if (!query) return true;
  const haystack = `${pkg.name} ${pkg.packageId} ${localizedDescription(pkg) || ""} ${localizedCategory(pkg)}`.toLowerCase();
  return haystack.includes(query);
}

function buildProfile() {
  return {
    version: 1,
    name: translate("profile.defaultName"),
    packages: rawCatalog.filter((pkg) => selectedProfilePackages.has(pkg.packageId)).map((pkg) => pkg.packageId),
  };
}

async function copyProfileJson() {
  if (selectedProfilePackages.size === 0 || !profileCopy) return;
  try {
    await navigator.clipboard.writeText(JSON.stringify(buildProfile(), null, 2));
    profileCopy.textContent = translate("common.copied");
    profileCopy.classList.add("copied");
  } catch {
    profileCopy.textContent = translate("common.copyFailed");
  }
  window.setTimeout(() => {
    profileCopy.textContent = translate("profile.copyJson");
    profileCopy.classList.remove("copied");
  }, 1400);
}

function downloadProfileJson() {
  if (selectedProfilePackages.size === 0) return;
  const blob = new Blob([`${JSON.stringify(buildProfile(), null, 2)}\n`], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "freshctl-profile.json";
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}

function openPackages(event) {
  event.preventDefault();
  if (!packageModal || !packageSearch) return;

  packageModal.classList.add("open");
  packageModal.setAttribute("aria-hidden", "false");
  document.body.classList.add("modal-open");
  renderPackages();

  window.setTimeout(() => packageSearch.focus(), 0);
}

function closePackages() {
  if (!packageModal) return;

  packageModal.classList.remove("open");
  packageModal.setAttribute("aria-hidden", "true");
  document.body.classList.remove("modal-open");
}

function openInstallSpotlight(event) {
  event.preventDefault();
  if (!installCard || !installSpotlight) return;

  closePackages();
  window.clearTimeout(installSpotlightTimer);
  installCard.scrollIntoView({ behavior: "smooth", block: "center" });

  window.setTimeout(() => {
    document.body.classList.add("install-spotlight-open");
    installSpotlight.classList.add("open");
  }, 180);

  installSpotlightTimer = window.setTimeout(closeInstallSpotlight, 2500);
}

function closeInstallSpotlight() {
  window.clearTimeout(installSpotlightTimer);
  document.body.classList.remove("install-spotlight-open");
  installSpotlight?.classList.remove("open");
}

function escapeHtml(value) {
  return value.replace(/[&<>"']/g, (char) => {
    const entities = {
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#039;",
    };
    return entities[char];
  });
}

packageOpeners.forEach((button) => button.addEventListener("click", openPackages));
packageClosers.forEach((button) => button.addEventListener("click", closePackages));
packageSearch?.addEventListener("input", renderPackages);
profileSearch?.addEventListener("input", renderProfileGenerator);
profilePackageList?.addEventListener("click", (event) => {
  const row = event.target.closest("[data-profile-package]");
  if (!row) return;
  const packageId = row.getAttribute("data-profile-package");
  if (!packageId) return;
  if (selectedProfilePackages.has(packageId)) {
    selectedProfilePackages.delete(packageId);
  } else {
    selectedProfilePackages.add(packageId);
  }
  renderProfileGenerator();
});
profileCopy?.addEventListener("click", copyProfileJson);
profileDownload?.addEventListener("click", downloadProfileJson);
languageButtons.forEach((button) => {
  button.addEventListener("click", () => applyLanguage(button.dataset.language, true));
});
installSpotlightLink?.addEventListener("click", openInstallSpotlight);
installSpotlight?.addEventListener("click", closeInstallSpotlight);
applyLanguage(currentLanguage);

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && packageModal?.classList.contains("open")) {
    closePackages();
  }
  if (event.key === "Escape" && installSpotlight?.classList.contains("open")) {
    closeInstallSpotlight();
  }
});
