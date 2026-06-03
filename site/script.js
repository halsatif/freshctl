const rawCatalog = (window.FRESHCTL_CATALOG || []).map((pkg) => ({ ...pkg, id: pkg.packageId }));
const packageCatalog = [...rawCatalog].sort((left, right) =>
  left.name.localeCompare(right.name, undefined, { sensitivity: "base" }),
);

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
let installSpotlightTimer;
const selectedProfilePackages = new Set();

copyButtons.forEach((button) => {
  button.addEventListener("click", async () => {
    const text = button.getAttribute("data-copy");
    if (!text) return;

    try {
      await navigator.clipboard.writeText(text);
      const previous = button.textContent;
      button.textContent = "Copied";
      button.classList.add("copied");

      window.setTimeout(() => {
        button.textContent = previous;
        button.classList.remove("copied");
      }, 1400);
    } catch {
      button.textContent = "Copy failed";
      window.setTimeout(() => {
        button.textContent = "Copy";
      }, 1400);
    }
  });
});

function renderPackages() {
  if (!packageList || !packageCount || !packageSearch) return;

  const query = packageSearch.value.trim().toLowerCase();
  const visible = packageCatalog.filter((pkg) => {
    const haystack = `${pkg.name} ${pkg.packageId} ${pkg.description} ${pkg.category} ${pkg.type}`.toLowerCase();
    return haystack.includes(query);
  });

  packageCount.textContent = `${visible.length} of ${packageCatalog.length} packages`;

  if (visible.length === 0) {
    packageList.innerHTML = '<div class="package-empty">No packages found.</div>';
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

  profileCount.textContent = `${selectedCount} selected • ${visible.length} shown`;
  profileOutput.textContent = hasSelection ? JSON.stringify(profile, null, 2) : "";
  if (profileHint) {
    profileHint.textContent = hasSelection
      ? "Save this file as freshctl-profile.json and import it with o in freshctl."
      : "Select packages to generate a profile.";
  }
  if (profileCopy) profileCopy.disabled = !hasSelection;
  if (profileDownload) profileDownload.disabled = !hasSelection;

  if (visible.length === 0) {
    profilePackageList.innerHTML = '<div class="package-empty">No packages found.</div>';
    return;
  }

  profilePackageList.innerHTML = visible
    .map((pkg) => {
      const selected = selectedProfilePackages.has(pkg.packageId);
      return `
        <button class="profile-package-row${selected ? " selected" : ""}" type="button" data-profile-package="${escapeHtml(pkg.packageId)}">
          <span>
            <strong>${escapeHtml(pkg.name)}</strong>
            <small>${escapeHtml(pkg.description || pkg.packageId)}</small>
          </span>
          <code>${selected ? "selected" : pkg.packageId}</code>
        </button>
      `;
    })
    .join("");
}

function matchesProfileQuery(pkg, query) {
  if (!query) return true;
  const haystack = `${pkg.name} ${pkg.packageId} ${pkg.description || ""}`.toLowerCase();
  return haystack.includes(query);
}

function buildProfile() {
  return {
    version: 1,
    name: "freshctl profile",
    packages: rawCatalog.filter((pkg) => selectedProfilePackages.has(pkg.packageId)).map((pkg) => pkg.packageId),
  };
}

async function copyProfileJson() {
  if (selectedProfilePackages.size === 0 || !profileCopy) return;
  const previous = profileCopy.textContent;
  try {
    await navigator.clipboard.writeText(JSON.stringify(buildProfile(), null, 2));
    profileCopy.textContent = "Copied";
    profileCopy.classList.add("copied");
  } catch {
    profileCopy.textContent = "Copy failed";
  }
  window.setTimeout(() => {
    profileCopy.textContent = previous;
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
installSpotlightLink?.addEventListener("click", openInstallSpotlight);
installSpotlight?.addEventListener("click", closeInstallSpotlight);
renderProfileGenerator();

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && packageModal?.classList.contains("open")) {
    closePackages();
  }
  if (event.key === "Escape" && installSpotlight?.classList.contains("open")) {
    closeInstallSpotlight();
  }
});
