# TaskFlow CI/CD Pipeline — Kelompok 6

Selamat datang di repository Kelompok 6 untuk mata kuliah Operasional Pengembang (DevOps). Project ini fokus pada implementasi CI/CD menggunakan **Jenkins** untuk aplikasi Backend berbasis Go.

---

## 📋 Daftar Isi
1. [Skenario 1: Bug Fix & Test](#-skenario-1-bug-fix--test)
2. [Skenario 2: Jenkins Setup (Detail)](#-skenario-2-jenkins-setup-detail)
3. [Konfigurasi Pipeline CI](#-konfigurasi-pipeline-ci)
4. [Perintah Pengembangan](#-perintah-pengembangan)

---

## ✅ Skenario 1: Bug Fix & Test
Tujuan: Memperbaiki 3 bug logika dan memastikan stabilitas kode melalui pengujian otomatis.

### 1. Daftar Perbaikan Bug
| Bug ID | Deskripsi Masalah | File Terkait | Solusi |
|---|---|---|---|
| **Bug #1** | *Integer Division* (Hasil stats selalu 0%) | `internal/service/service.go` | Casting `float64()` pada variabel `completed` dan `len(tasks)`. |
| **Bug #2** | Filter Status Terbalik (`!=` vs `==`) | `internal/repository/memory.go` & `postgres.go` | Mengubah operator logika menjadi `==` untuk filter status yang benar. |
| **Bug #3** | Validasi Priority `"urgent"` Ilegal | `internal/validator/validator.go` | Menghapus `"urgent"` dari map validasi (hanya `low, medium, high`). |

### 2. Verifikasi Hasil
Semua bug telah diverifikasi lulus test dengan cakupan pengujian:
*   **Unit Test**: `PASS`
*   **Race Detector**: `PASS` (Tidak ada data race)
*   **Coverage**: **83.1%** (Melebihi target 75%)

---

## 🚀 Skenario 2: Jenkins Setup (Detail)
Gunakan langkah-langkah di bawah ini untuk menjalankan Jenkins di laptop masing-masing tanpa memerlukan VPS.

### 1. Jalankan Jenkins via Docker
Buka Terminal/PowerShell, jalankan perintah ini (sesuaikan jika menggunakan Windows/CMD):
```bash
# Untuk PowerShell (Windows)
docker run -d --name jenkins-local `
  -p 8080:8080 -p 50000:50000 `
  -v jenkins_home:/var/jenkins_home `
  -v /var/run/docker.sock:/var/run/docker.sock `
  jenkins/jenkins:lts
```

### 2. Ambil Initial Admin Password
Setelah container berjalan (tunggu ~1 menit), ambil password untuk login pertama kali:
```bash
docker exec jenkins-local cat /var/jenkins_home/secrets/initialAdminPassword
```
*Copy kode yang muncul dan paste di `http://localhost:8080`.*

### 3. Instalasi Plugins & Wizard
1.  Pilih **"Install Suggested Plugins"** pada layar awal.
2.  Setelah masuk ke Dashboard, pergi ke: **Manage Jenkins** -> **Plugins** -> **Available Plugins**.
3.  Cari dan centang plugin berikut:
    *   **Go Plugin**
    *   **Docker Pipeline**
4.  Klik **Install without restart**.

### 4. Konfigurasi Go (Global Tool)
Agar Jenkins bisa menjalankan perintah `go`, kita harus mendaftarkan tools-nya:
1.  Buka **Manage Jenkins** -> **Tools**.
2.  Scroll ke bagian **Go**, klik **Add Go**.
3.  Nama: `go1.22` (PENTING: Gunakan nama ini agar sesuai Jenkinsfile).
4.  Centang **Install automatically**.
5.  Pilih versi terbaru di seri 1.22 (misal: **Go 1.22.12**).
6.  Klik **Save**.

---

## 🛠 Konfigurasi Pipeline CI
Langkah terakhir untuk menghubungkan kode ke Jenkins:

1.  Klik **New Item** pada sidebar Jenkins.
2.  Masukkan nama: `TaskFlow-CI-CD` -> Pilih **Pipeline** -> Klik **OK**.
3.  Scroll ke bagian **Pipeline**:
    *   Definition: **Pipeline script from SCM**
    *   SCM: **Git**
    *   Repository URL: *URL Repo GitHub kamu (atau path lokal)*
    *   Script Path: `pertemuan-09-cicd/pbl-taskflow-go/Jenkinsfile`
4.  Klik **Save**.
5.  Klik **Build Now** untuk menjalankan pipeline pertama kali.

---

## 💻 Perintah Pengembangan
Gunakan perintah ini saat bekerja di folder `pertemuan-09-cicd/pbl-taskflow-go/`:

| Perintah | Fungsi |
|---|---|
| `go test ./...` | Menjalankan seluruh test. |
| `go test -race ./...` | Mengecek adanya data race condition. |
| `make build` | Membuat file binary aplikasi. |
| `make up` | Menjalankan Full Stack (App + Postgres). |
| `make clean` | Menghapus file temporary/build. |

---
**Status Project**: 🏗️ S2 In Progress | 🏁 S1 Completed.
