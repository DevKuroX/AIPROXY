# Install — aiproxy-autonomous-loop

Skill ini custom skill untuk Codex (CLI/IDE) yang override mode planning bawaan
saat ngerjain AIPROXY frontend migration di mode idle.

## Lokasi Install (pilih salah satu)

### Opsi A — Per-repo (rekomendasi)

Drop folder skill ini ke repo AIPROXY supaya semua kontributor / agent dapat
auto-load:

```
AIPROXY/
└── .codex/
    └── skills/
        └── aiproxy-autonomous-loop/
            ├── SKILL.md
            ├── PROMPT_TEMPLATES.md
            └── INSTALL.md
```

Commit ke branch `main` atau ke branch tooling. Tambahkan ke `.gitignore`
**JANGAN** — skill perlu ke-share.

### Opsi B — Global per-user

```
~/.codex/skills/aiproxy-autonomous-loop/
```

Cocok kalau lu satu-satunya yang pakai dan repo lain juga ngikut skill yang
sama. Tapi untuk AIPROXY (yang banyak agent), pilih Opsi A.

## Verifikasi Install

1. Buka Codex CLI/IDE di repo AIPROXY.
2. Ketik trigger phrase, contoh:

   ```
   Aktifkan skill aiproxy-autonomous-loop. Eksekusi semua task Phase 2 sampai selesai.
   ```

3. Output pertama Codex HARUS:

   ```
   [aiproxy-autonomous-loop] active — Codex built-in planning disabled.
   ```

   Kalau Codex malah jalanin `/plan` bawaan atau nanya konfirmasi tiap task,
   berarti skill belum ke-load. Check path & nama file (`SKILL.md` exact).

## Catatan untuk Codex Idle Mode

- Idle mode Codex biasanya jalan tanpa user di depan layar. Skill ini dirancang
  untuk itu: NO interactive prompts, NO mid-loop confirmation.
- Pastikan environment punya:
  - Git config (`user.name`, `user.email`) terset
  - Branch `phase/<N>-<slug>` udah di-checkout SEBELUM trigger skill
  - `npm install` udah jalan di `frontend/`
  - Backend Go bisa diakses di `http://localhost:1432` kalau task perlu HAR
    replay atau live verification

## Update / Edit Skill

Edit `SKILL.md` langsung. Jangan rename file. Versi diatur lewat git history.
Setelah edit:

```bash
git add .codex/skills/aiproxy-autonomous-loop/
git commit -m "chore(skill): update aiproxy-autonomous-loop"
```

— END —
