#!/usr/bin/env bash
set -e

# Konfigurasi
BIN_NAME="gibrun"
INSTALL_DIR="/usr/local/bin"
POLKIT_RULE="/etc/polkit-1/rules.d/49-gibrun.rules"
GROUP="gibrun"
REPO="Rauliqbal/gibrun-tui"
VERSION="v0.1.0"

# Warna untuk output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}⚡ Menyiapkan instalasi $BIN_NAME...${NC}"

# 1. Self-Elevate (Auto-Sudo)
if [ "$EUID" -ne 0 ]; then
    echo -e "${YELLOW}🔐 Memerlukan hak akses root. Meminta sudo...${NC}"
    if [ "$0" = "sh" ] || [ "$0" = "bash" ] || [ "$0" = "-" ]; then
        tmp_script=$(mktemp)
        cat > "$tmp_script"
        sudo bash "$tmp_script" "$@"
        rm "$tmp_script"
    else
        exec sudo bash "$0" "$@"
    fi
    exit $?
fi

# 2. Deteksi Arsitektur & Download (Jika binary lokal tidak ada)
if [ ! -f "./$BIN_NAME" ]; then
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)  SUFFIX="amd64" ;;
        aarch64) SUFFIX="arm64" ;;
        *)       echo -e "${RED}❌ Arsitektur $ARCH tidak didukung.${NC}"; exit 1 ;;
    esac

    echo -e "${BLUE}📥 Mengunduh binary untuk $ARCH...${NC}"
    # Pastikan nama file di GitHub Release sesuai, misal: gibrun-linux-amd64
    URL="https://github.com/$REPO/releases/download/$VERSION/$BIN_NAME"
    
    if ! curl -fsSL "$URL" -o "$BIN_NAME"; then
        echo -e "${RED}❌ Gagal mengunduh binary. Pastikan Release $VERSION sudah dipublish.${NC}"
        exit 1
    fi
    chmod +x "$BIN_NAME"
fi

# 3. Eksekusi Instalasi
echo -e "${BLUE}📦 Menyalin binary ke $INSTALL_DIR...${NC}"
install -m 755 "$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"

# 4. Grup & Izin (Polkit)
echo -e "${BLUE}👥 Mengatur grup sistem & izin Polkit...${NC}"
if ! getent group "$GROUP" >/dev/null; then
    groupadd "$GROUP"
fi

cat > "$POLKIT_RULE" <<EOF
polkit.addRule(function(action, subject) {
    if (action.id === "org.freedesktop.systemd1.manage-units" && subject.isInGroup("$GROUP")) {
        return polkit.Result.YES;
    }
});
EOF
chmod 644 "$POLKIT_RULE"

# 5. Konfigurasi User
CURRENT_USER=${SUDO_USER:-$(whoami)}
usermod -aG "$GROUP" "$CURRENT_USER"

# 6. Selesai (Final Copywriting)
echo -e "\n${GREEN}✨ Instalasi Berhasil!${NC}"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "👤 User [${YELLOW}$CURRENT_USER${NC}] telah ditambahkan ke grup [${YELLOW}$GROUP${NC}]."
echo -e "⚙️  Hak akses tanpa password telah dikonfigurasi."
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "\n${BLUE}👉 Langkah terakhir:${NC}"
echo -e "Jalankan perintah ini agar perubahan grup aktif tanpa logout:"
echo -e "${YELLOW}newgrp $GROUP${NC}"
echo -e "\nLalu jalankan aplikasi dengan mengetik: ${GREEN}$BIN_NAME${NC}"