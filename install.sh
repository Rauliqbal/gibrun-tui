#!/usr/bin/env bash

set -e

# === Konfigurasi ===
BIN_NAME="gibrun"
INSTALL_DIR="/usr/local/bin"
POLKIT_RULE="/etc/polkit-1/rules.d/49-gibrun.rules"
GROUP="gibrun"
REPO="Rauliqbal/gibrun-tui"
VERSION="v0.1.0" # Ganti sesuai versi rilis Anda

# === Warna untuk UI ===
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color
BOLD='\033[1m'

echo -e "${BLUE}${BOLD}⚡ GibRun Installer${NC}"
echo -e "${BLUE}============================${NC}"

# 1. Self-Elevate (Auto-Sudo) - Kompatibel dengan Pipe & Local
if [ "$EUID" -ne 0 ]; then
    echo -e "${YELLOW}🔐 Memerlukan hak akses root. Meminta sudo...${NC}"
    if [ ! -f "$0" ]; then
        # Jika dijalankan via pipe (curl | bash), simpan sementara
        TMP_FILE=$(mktemp /tmp/gibrun-install.XXXXXX.sh)
        cat > "$TMP_FILE"
        sudo bash "$TMP_FILE" "$@"
        rm -f "$TMP_FILE"
    else
        # Jika dijalankan sebagai file lokal (./install.sh)
        exec sudo bash "$0" "$@"
    fi
    exit $?
fi

# 2. Deteksi Arsitektur & Download Binary
if [ ! -f "./$BIN_NAME" ]; then
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)  SUFFIX="amd64" ;;
        aarch64) SUFFIX="arm64" ;;
        *)       echo -e "${RED}❌ Arsitektur $ARCH tidak didukung.${NC}"; exit 1 ;;
    esac

    echo -e "${BLUE}📥 Mengunduh binary [$ARCH] dari GitHub...${NC}"
    # Link mengarah ke release assets
    URL="https://github.com/$REPO/releases/download/$VERSION/$BIN_NAME"
    
    if ! curl -fsSL "$URL" -o "$BIN_NAME"; then
        echo -e "${RED}❌ Gagal mengunduh binary.${NC}"
        echo -e "Pastikan Release ${YELLOW}$VERSION${NC} sudah dipublish di GitHub."
        exit 1
    fi
    chmod +x "$BIN_NAME"
fi

# 3. Instalasi Binary
echo -e "${BLUE}📦 Menyalin binary ke $INSTALL_DIR...${NC}"
install -m 755 "$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"

# 4. Konfigurasi Grup & Polkit
echo -e "${BLUE}👥 Mengatur grup sistem & izin Polkit...${NC}"
if ! getent group "$GROUP" >/dev/null; then
    groupadd "$GROUP"
    echo -e "  - Grup ${GREEN}$GROUP${NC} berhasil dibuat"
fi

# Tulis rule Polkit agar manage service tidak minta password
cat > "$POLKIT_RULE" <<EOF
polkit.addRule(function(action, subject) {
    if (
        action.id === "org.freedesktop.systemd1.manage-units" &&
        subject.isInGroup("$GROUP")
    ) {
        return polkit.Result.YES;
    }
});
EOF
chmod 644 "$POLKIT_RULE"

# 5. Daftarkan User ke Grup
CURRENT_USER=${SUDO_USER:-$(whoami)}
if [ "$CURRENT_USER" != "root" ]; then
    usermod -aG "$GROUP" "$CURRENT_USER"
fi

# 6. Selesai
echo -e "\n${GREEN}${BOLD}✨ Instalasi Berhasil Selesai!${NC}"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "👤 User: ${YELLOW}$CURRENT_USER${NC}"
echo -e "⚙️  Grup: ${YELLOW}$GROUP${NC} (Akses Systemd diaktifkan)"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "\n${BLUE}${BOLD}👉 Langkah Terakhir:${NC}"
echo -e "Agar perubahan grup aktif tanpa logout, jalankan perintah ini:"
echo -e "${GREEN}newgrp $GROUP${NC}"
echo -e "\nLalu jalankan aplikasi dengan mengetik: ${GREEN}$BIN_NAME${NC}"