# ⚡ GibRun - Lightweight service control

**GibRun** adalah Terminal User Interface (TUI) yang ringan dan tangguh untuk mengelola layanan `systemd` pada sistem Linux. Didesain untuk efisiensi, GibRun memungkinkan Anda memantau log secara realtime, menjalankan, menghentikan, dan menginstal layanan hanya dengan satu tombol tanpa perlu mengetik perintah `systemctl` berulang kali.

---

![GibRun Preview](/preview.png)


## ✨ Fitur Utama

* **Realtime Log Streaming**: Pantau aktivitas layanan langsung dari antarmuka TUI.
* **Quick Control**: Start, Stop, dan Restart layanan dengan satu tombol.
* **Auto-Detection**: Mendeteksi distro Linux, port usage, dan uptime layanan secara otomatis.
* **Smart Installer**: Mengonfigurasi grup sistem dan izin Polkit secara otomatis agar Anda bisa mengelola layanan tanpa perlu mengetik password terus-menerus.
* **XDG Compliant**: Menyimpan konfigurasi secara rapi di `~/.config/gibrun/`.


## 🚀 Instalasi

Jalankan perintah berikut untuk mengunduh dan memasang GibRun secara otomatis:

```bash
curl -fsSL https://raw.githubusercontent.com/Rauliqbal/gibrun-tui/main/install.sh -o install.sh
bash install.sh
rm install.sh
```

**Penting**: Setelah instalasi selesai, jalankan perintah newgrp gibrun atau restart terminal Anda agar izin grup baru aktif.


## ⚙️ Konfigurasi
Anda dapat menambahkan atau mengubah daftar layanan yang dipantau dengan mengedit file YAML berikut:

```bash
nano ~/.config/gibrun/services.yml
```

contoh isi `services.yml`
```yml
nginx:
  service_name: nginx
  description: Web Server
docker:
  service_name: docker
  description: Docker Engine
```

## 🔒 Privasi
GibRun sangat menghargai privasi Anda:

- **Tanpa Koleksi Data**: GibRun tidak mengumpulkan, menyimpan, atau mengirim data apa pun ke server luar.
- **Offline First**: Seluruh proses monitoring dan manajemen dilakukan secara lokal di mesin Anda.


## 📄 Lisensi
Proyek ini berada di bawah lisensi [MIT](LICENSE)

Dibuat karena [Rauliqbal](https://rauliqbal.my.id) ribet aja kalo gonta ganti distro 