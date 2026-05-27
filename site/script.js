const packageCatalog = [
  ["Google Chrome", "googlechrome"],
  ["Opera", "opera"],
  ["Opera GX", "opera-gx"],
  ["Mozilla Firefox", "firefox"],
  ["Waterfox", "waterfox"],
  ["Microsoft Edge", "microsoft-edge"],
  ["Brave Browser", "brave"],
  ["Vivaldi", "vivaldi"],
  ["Yandex Browser", "yandex-browser"],
  ["Tor Browser", "tor-browser"],
  ["LibreWolf", "librewolf"],
  ["Zen Browser", "zen-browser"],
  ["Telegram Desktop", "telegram"],
  ["Signal", "signal"],
  ["Element", "element-desktop"],
  ["Zoom", "zoom"],
  ["Microsoft Teams", "microsoft-teams"],
  ["Dorion", "dorion"],
  ["Visual Studio Code", "vscode"],
  ["Zed", "zed-editor"],
  ["Sublime Text", "sublimetext4"],
  ["Neovim", "neovim"],
  ["Helix", "helix"],
  ["JetBrains Toolbox", "jetbrainstoolbox"],
  ["IntelliJ IDEA Community", "intellijidea-community"],
  ["PyCharm Community", "pycharm-community"],
  ["Android Studio", "androidstudio"],
  ["Git", "git"],
  [".NET Runtime 10", "dotnet-10.0-runtime"],
  [".NET Runtime 9", "dotnet-9.0-runtime"],
  [".NET Runtime 8", "dotnet-8.0-runtime"],
  [".NET Runtime 7", "dotnet-7.0-runtime"],
  [".NET Runtime 6", "dotnet-6.0-runtime"],
  [".NET Runtime 5", "dotnet-5.0-runtime"],
  [".NET SDK 10", "dotnet-10.0-sdk"],
  [".NET SDK 9", "dotnet-9.0-sdk"],
  [".NET SDK 8", "dotnet-8.0-sdk"],
  [".NET SDK 7", "dotnet-7.0-sdk"],
  [".NET SDK 6", "dotnet-6.0-sdk"],
  [".NET SDK 5", "dotnet-5.0-sdk"],
  ["JDK 25 (Adoptium)", "temurin25"],
  ["JDK 21 (Adoptium)", "temurin21"],
  ["JDK 17 (Adoptium)", "temurin17"],
  ["JDK 11 (Adoptium)", "temurin11"],
  ["JDK 8 (Adoptium)", "temurin8"],
  ["JRE 25 (Adoptium)", "temurin25jre"],
  ["JRE 21 (Adoptium)", "temurin21jre"],
  ["JRE 17 (Adoptium)", "temurin17jre"],
  ["JRE 11 (Adoptium)", "temurin11jre"],
  ["JRE 8 (Adoptium)", "temurin8jre"],
  ["Node.js LTS", "nodejs-lts"],
  ["Python 3", "python"],
  ["Go", "golang"],
  ["Rustup", "rustup.install"],
  ["LLVM", "llvm"],
  ["MinGW", "mingw"],
  ["CMake", "cmake"],
  ["VC++ Redist 2010 x86/x64", "vcredist2010"],
  ["VC++ Redist 2012 x86/x64", "vcredist2012"],
  ["VC++ Redist 2013 x86/x64", "vcredist2013"],
  ["VC++ Redist 2015-2022 x86/x64", "vcredist140"],
  ["Windows Terminal", "microsoft-windows-terminal"],
  ["PowerShell 7", "powershell-core"],
  ["WezTerm", "wezterm"],
  ["Fastfetch", "fastfetch"],
  ["FZF", "fzf"],
  ["ripgrep", "ripgrep"],
  ["Codex", "codex"],
  ["Docker Desktop", "docker-desktop"],
  ["Podman Desktop", "podman-desktop"],
  ["Postman", "postman"],
  ["Bruno", "bruno"],
  ["Insomnia", "insomnia-rest-api-client"],
  ["DBeaver", "dbeaver"],
  ["PostgreSQL", "postgresql"],
  ["MySQL", "mysql"],
  ["MongoDB Compass", "mongodb-compass"],
  ["iTunes", "itunes"],
  ["VLC", "vlc"],
  ["AIMP", "aimp"],
  ["foobar2000", "foobar2000"],
  ["Winamp", "winamp"],
  ["MusicBee", "musicbee"],
  ["Audacious", "audacious"],
  ["Audacity", "audacity"],
  ["K-Lite Codecs", "k-litecodecpackfull"],
  ["GOM", "gom-player"],
  ["mpv", "mpvio"],
  ["Spotify", "spotify"],
  ["OBS Studio", "obs-studio"],
  ["Kdenlive", "kdenlive"],
  ["HandBrake", "handbrake"],
  ["yt-dlp", "yt-dlp"],
  ["FFmpeg", "ffmpeg"],
  ["Krita", "krita"],
  ["Blender", "blender"],
  ["Paint.NET", "paint.net"],
  ["GIMP", "gimp"],
  ["IrfanView", "irfanview"],
  ["XnView", "xnview"],
  ["Inkscape", "inkscape"],
  ["FastStone Image Viewer", "fsviewer"],
  ["Greenshot", "greenshot"],
  ["Lightshot", "lightshot"],
  ["ImageGlass", "imageglass"],
  ["ShareX", "sharex"],
  ["ScreenToGif", "screentogif"],
  ["Flameshot", "flameshot"],
  ["ImgBurn", "imgburn"],
  ["CDBurnerXP", "cdburnerxp"],
  ["InfraRecorder", "infrarecorder"],
  ["Steam", "steam"],
  ["Epic Games Launcher", "epicgameslauncher"],
  ["Heroic Games Launcher", "heroic-games-launcher"],
  ["Prism Launcher", "prismlauncher"],
  ["Discord", "discord"],
  ["Parsec", "parsec"],
  ["Moonlight", "moonlight"],
  ["Sunshine", "sunshine"],
  ["MSI Afterburner", "msiafterburner"],
  ["AnyDesk", "anydesk"],
  ["TeamViewer", "teamviewer"],
  ["RealVNC Server", "vnc-connect"],
  ["RealVNC Viewer", "vnc-viewer"],
  ["TightVNC", "tightvnc"],
  ["RustDesk", "rustdesk"],
  ["Barrier", "barrier"],
  ["scrcpy", "scrcpy"],
  ["ADB Platform Tools", "adb"],
  ["Everything", "everything"],
  ["TeraCopy", "teracopy"],
  ["Revo Uninstaller", "revo-uninstaller"],
  ["Launchy", "launchy"],
  ["WinDirStat", "windirstat"],
  ["WizTree", "wiztree"],
  ["Glary Utilities", "glaryutilities-free"],
  ["Open-Shell", "open-shell"],
  ["CCleaner", "ccleaner"],
  ["PowerToys", "powertoys"],
  ["Google Earth", "googleearthpro"],
  ["AutoHotkey", "autohotkey"],
  ["Ventoy", "ventoy"],
  ["Bulk Crap Uninstaller", "bulk-crap-uninstaller"],
  ["HWiNFO64", "hwinfo"],
  ["HWMonitor", "hwmonitor"],
  ["CPU-Z", "cpu-z"],
  ["GPU-Z", "gpu-z"],
  ["System Informer", "systeminformer"],
  ["Process Explorer", "procexp"],
  ["Autoruns", "autoruns"],
  ["TreeSize Free", "treesizefree"],
  ["EarTrumpet", "eartrumpet"],
  ["StartAllBack", "startallback"],
  ["TranslucentTB", "translucenttb"],
  ["F.lux", "flux"],
  ["Twinkle Tray", "twinkle-tray"],
  ["UniGetUI", "unigetui"],
  ["7-Zip", "7zip"],
  ["WinRAR", "winrar"],
  ["PeaZip", "peazip"],
  ["Bitwarden", "bitwarden"],
  ["KeePass 2", "keepass"],
  ["Malwarebytes", "malwarebytes"],
  ["VeraCrypt", "veracrypt"],
  ["BleachBit", "bleachbit"],
  ["SimpleWall", "simplewall"],
  ["FileZilla", "filezilla"],
  ["WinSCP", "winscp"],
  ["PuTTY", "putty"],
  ["qBittorrent", "qbittorrent"],
  ["Tailscale", "tailscale"],
  ["WireGuard", "wireguard"],
  ["ZeroTier", "zerotier-one"],
  ["Wireshark", "wireshark"],
  ["Nmap", "nmap"],
  ["Syncthing", "syncthing"],
  ["LocalSend", "localsend"],
  ["Dropbox", "dropbox"],
  ["Google Drive", "googledrive"],
  ["LibreOffice", "libreoffice-fresh"],
  ["OpenOffice", "openoffice"],
  ["Foxit Reader", "foxitreader"],
  ["Evernote", "evernote"],
  ["OnlyOffice", "onlyoffice"],
  ["SumatraPDF", "sumatrapdf"],
  ["Claude", "claude"],
  ["Notepad++", "notepadplusplus"],
  ["Cursor", "cursoride"],
  ["WinMerge", "winmerge"],
  ["balenaEtcher", "etcher"],
  ["VirtualBox", "virtualbox"],
]
  .map(([name, id]) => ({ name, id }))
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
    const haystack = `${pkg.name} ${pkg.id}`.toLowerCase();
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
          <code>${escapeHtml(pkg.id)}</code>
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
