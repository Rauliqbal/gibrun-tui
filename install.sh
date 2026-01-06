#!/usr/bin/env bash

# Keluar jika ada error
set -e

# === 1. Konfigurasi Utama ===
readonly REPO="Rauliqbal/gibrun-tui"
readonly VERSION="v0.1.0"
readonly BIN_NAME="gibrun"
readonly INSTALL_DIR="/usr/local/bin"
readonly POLKIT_RULE="/etc/polkit-1/rules.d/49-gibrun.rules"
readonly GROUP="gibrun"

# URL Mentah (Raw) untuk download file pendukung
readonly RAW_URL="https://raw.githubusercontent.com/$REPO/main"

# === 2. Warna untuk UI ===
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'
BOLD='\033[1m'

# === 3. Header ===
clear
echo -e "${BLUE}${BOLD}⚡ GibRun Installer (v$VERSION)${NC}"
echo -e "${BLUE}======================================${NC}"

# === 4. Pengecekan Root (Auto-Sudo) ===
if [ "$EUID" -ne 0 ]; then
    echo -e "${YELLOW}🔐 Memerlukan hak akses root. Meminta sudo...${NC}"
    exec sudo -E bash "$0" "$@"
    exit $?
fi

# Variabel User (Penting untuk konfigurasi lokal)
REAL_USER=${SUDO_USER:-$(whoami)}
USER_HOME=$(eval echo "~$REAL_USER")
CONFIG_DIR="$USER_HOME/.config/gibrun"

# === 5. Deteksi Arsitektur & Download Binary ===
if [ ! -f "./$BIN_NAME" ]; then
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)  SUFFIX="amd64" ;;
        aarch64) SUFFIX="arm64" ;;
        *)       echo -e "${RED}❌ Arsitektur $ARCH tidak didukung.${NC}"; exit 1 ;;
    esac

    echo -e "${BLUE}📥 Mengunduh binary [$ARCH] dari GitHub...${NC}"
    URL="https://github.com/$REPO/releases/download/$VERSION/$BIN_NAME"
    
    if ! curl -fsSL "$URL" -o "$BIN_NAME"; then
        echo -e "${RED}❌ Gagal mengunduh binary.${NC}"
        exit 1
    fi
    chmod +x "$BIN_NAME"
fi

# === 6. Instalasi Binary ke Sistem ===
echo -e "${BLUE}📦 Menyalin binary ke $INSTALL_DIR...${NC}"
install -m 755 "$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"

# === 7. Setup Konfigurasi (~/.config/gibrun) ===
echo -e "${BLUE}⚙️  Menyiapkan folder konfigurasi di $CONFIG_DIR...${NC}"
mkdir -p "$CONFIG_DIR"

# LOGIKA BARU: Ambil config dari GitHub jika tidak ada di lokal
if [ ! -f "$CONFIG_DIR/services.yml" ]; then
    echo -e "${BLUE}📄 Mengunduh default services.yml dari GitHub...${NC}"
    # Pastikan path internal/config/services.yml sesuai dengan struktur repo Anda
    if ! curl -fsSL "$RAW_URL/internal/config/services.yml" -o "$CONFIG_DIR/services.yml"; then
        # Jika path di atas salah, coba path alternatif (root folder)
        curl -fsSL "$RAW_URL/services.yml" -o "$CONFIG_DIR/services.yml" || echo -e "${RED}⚠️  Gagal mengunduh config.${NC}"
    fi
else
    echo -e "${YELLOW}ℹ️  File services.yml sudah ada, melewati pengunduhan.${NC}"
fi

# Kembalikan kepemilikan folder config ke user biasa
chown -R "$REAL_USER:$REAL_USER" "$CONFIG_DIR"
chmod 644 "$CONFIG_DIR/services.yml"

# === 8. Setup Grup & Izin Polkit ===
echo -e "${BLUE}🔐 Mengatur izin Polkit & Grup...${NC}"
if ! getent group "$GROUP" >/dev/null; then
    groupadd "$GROUP"
fi

usermod -aG "$GROUP" "$REAL_USER"

cat > "$POLKIT_RULE" <<EOF
polkit.addRule(function(action, subject) {
    if (action.id === "org.freedesktop.systemd1.manage-units" && subject.isInGroup("$GROUP")) {
        return polkit.Result.YES;
    }
});
EOF
chmod 644 "$POLKIT_RULE"

# === 9. Selesai ===
echo -e "\n${GREEN}${BOLD}✨ Instalasi Berhasil Selesai!${NC}"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "📂 Config: ${YELLOW}$CONFIG_DIR/services.yml${NC}"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "\n${BLUE}${BOLD}👉 Langkah Terakhir:${NC}"
echo -e "${YELLOW}newgrp $GROUP && $BIN_NAME${NC}\n"