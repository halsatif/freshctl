# Changelog

## v0.1.1

Polish release focused on catalog reliability, installer trust, and clean-system testing.

### Changed

- Removed TeamSpeak from the catalog. The Chocolatey package `teamspeak` currently fails during install on clean systems because the upstream TeamSpeak download URL returns `403 Forbidden`; Chocolatey also flags it as likely broken for FOSS users.
- Removed legacy Visual C++ Redistributables `vcredist2005` and `vcredist2008` after clean Windows testing showed unreliable MSI/runtime assembly failures on modern Windows.
- Renamed `VC++ Redist 2015-2026` to `VC++ Redist 2015-2022` to match the expected Microsoft runtime branding more closely while keeping the `vcredist140` package.
- Removed packages that are not reliable default unattended installs: `faceit`, `nvidia-broadcast`, `vmwareworkstation`, `protonvpn`, and `rufus`.
- Removed Yandex Browser from the default catalog after Windows Sandbox smoke testing showed failed installs while other browsers completed successfully.
- Kept package catalog validation stricter: Chocolatey package ID existence is not enough, because package install scripts can still fail at runtime when upstream URLs, checksums, or installers change.
