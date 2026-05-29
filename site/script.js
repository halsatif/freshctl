const packageCatalog = (window.FRESHCTL_CATALOG || [])
  .map((pkg) => ({ ...pkg, id: pkg.packageId }))
  .sort((left, right) => left.name.localeCompare(right.name, undefined, { sensitivity: "base" }));

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
let installSpotlightTimer;

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
installSpotlightLink?.addEventListener("click", openInstallSpotlight);
installSpotlight?.addEventListener("click", closeInstallSpotlight);

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && packageModal?.classList.contains("open")) {
    closePackages();
  }
  if (event.key === "Escape" && installSpotlight?.classList.contains("open")) {
    closeInstallSpotlight();
  }
});
