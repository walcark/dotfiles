#!/usr/bin/env bash
set -Eeuo pipefail

function download_keepassxc() {
    local version="$1"
    local app="KeePassXC-${version}-x86_64.AppImage"
    local url="https://github.com/keepassxreboot/keepassxc/releases/download/${version}/${app}"

    # Dossiers d'installation
    local base_dir="$HOME/.local/share/keepassxc"
    local extract_dir="keepassxc_extract"

    echo "[INFO] Téléchargement KeePassXC $version..."
    curl -L -o "$app" "$url"

    echo "[INFO] Extraction de l'AppImage..."
    chmod +x "$app"
    "./$app" --appimage-extract

    echo "[INFO] Suppression des fichiers temporaires..."
    rm -f "$app"

    echo "[INFO] Renommage squashfs-root → keepassxc_extract..."
    rm -rf "$extract_dir" || true
    mv squashfs-root "$extract_dir"

    echo "[INFO] Déplacement dans ~/.local/share/keepassxc..."
    rm -rf "$base_dir" || true
    mkdir -p "$base_dir"
    mv "$extract_dir" "$base_dir/"

    echo "[INFO] Création des liens symboliques..."
    ln -sf "$base_dir/keepassxc_extract/usr/bin/keepassxc-cli" "$base_dir/keepassxc-cli"
    ln -sf "$base_dir/keepassxc_extract/usr/bin/keepassxc" "$base_dir/keepassxc"

    echo ""
    echo "[SUCCESS] Installation terminée."
    echo "CLI  : $base_dir/keepassxc-cli"
    echo "GUI  : $base_dir/keepassxc"
}

main() {
    download_keepassxc "2.6.6"
}

main "$@"

