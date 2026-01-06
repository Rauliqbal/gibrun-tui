#!/usr/bin/env bash
set -e

BIN_NAME="gibrun"
INSTALL_DIR="/usr/local/bin"
POLKIT_RULE="/etc/polkit-1/rules.d/49-gibrun.rules"
GROUP="gibrun"

echo "⚡ Installing gibrun..."

# ===== Check root =====
if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run installer as root"
  echo "   sudo sh install.sh"
  exit 1
fi

# ===== Create group =====
if ! getent group "$GROUP" >/dev/null; then
  echo "➕ Creating group '$GROUP'"
  groupadd "$GROUP"
else
  echo "✔ Group '$GROUP' already exists"
fi

# ===== Install binary =====
if [ ! -f "./$BIN_NAME" ]; then
  echo "❌ Binary '$BIN_NAME' not found"
  echo "   Run: go build -o gibrun"
  exit 1
fi

echo "📦 Installing binary → $INSTALL_DIR/$BIN_NAME"
install -m 755 "$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"

# ===== Polkit rule =====
echo "🔐 Installing polkit rule"
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

# ===== User info =====
CURRENT_USER=${SUDO_USER:-$(whoami)}

echo ""
echo "✅ Installation complete!"
echo ""
echo "👤 Add user to group:"
echo "   sudo usermod -aG $GROUP $CURRENT_USER"
echo ""
echo "🔁 Then logout/login to apply group change"
echo ""
echo "🚀 Run: gibrun"
