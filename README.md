# TaskFlow CI/CD Pipeline — Kelompok 6

Selamat datang di repository Kelompok 6 untuk mata kuliah Operasional Pengembang (DevOps).

---

## 📋 Daftar Isi
1. [Skenario 1: Bug Fix & Test](#-skenario-1-bug-fix--test)
2. [Skenario 2: Jenkins Setup (One-Click Setup)](#-skenario-2-jenkins-setup-one-click-setup)
3. [Konfigurasi Pipeline CI](#-konfigurasi-pipeline-ci)
4. [Perintah Pengembangan](#-perintah-pengembangan)

---

## ✅ Skenario 1: Bug Fix & Test
Tujuan: Memperbaiki 3 bug logika dan memastikan stabilitas kode.

| Bug ID | Masalah | Solusi |
|---|---|---|
| **Bug #1** | *Integer Division* | Casting `float64()` di `service.go`. |
| **Bug #2** | Filter Status Terbalik | Ubah `!=` menjadi `==` di repository. |
| **Bug #3** | Priority `"urgent"` | Hapus `"urgent"` dari map validasi. |

**Status**: `PASS` | **Coverage**: `83.1%`.

---

## 🚀 Skenario 2: Jenkins Setup (One-Click Setup)
Gunakan cara ini agar kamu tidak perlu setup Jenkins dari nol. Jenkins ini sudah otomatis terinstall plugin Go dan konfigurasi `go1.22`.

### 1. Jalankan Jenkins Otomatis
Buka Terminal di root project, jalankan:
```bash
docker compose -f jenkins-infra/docker-compose.yml up -d
```

### 2. Login ke Jenkins
*   Buka: `http://localhost:8080`
*   **Username**: `admin`
*   **Password**: `admin` (Sudah diatur otomatis, tidak perlu cari password di log).
*   *Setup wizard akan dilewati otomatis, semua plugin sudah terpasang.*

---

## 🛠 Konfigurasi Pipeline CI
1.  Klik **New Item** -> Nama: `TaskFlow-CI` -> **Pipeline** -> **OK**.
2.  Bagian **Pipeline**:
    *   Definition: **Pipeline script from SCM**
    *   SCM: **Git**
    *   Repository URL: `https://github.com/Neinstat/Devops-CI-CD-Kel-6-Jenkins`
    *   Script Path: `pertemuan-09-cicd/pbl-taskflow-go/Jenkinsfile`
3.  Klik **Save** & **Build Now**.

---

## 💻 Perintah Pengembangan
| Perintah | Fungsi |
|---|---|
| `make build` | Kompilasi aplikasi (Lokal). |
| `make up` | Jalankan Full Stack (App + Postgres). |

---
**Status Project**: 🏗️ S2 In Progress | 🏁 S1 Completed
