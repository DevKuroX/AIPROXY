# Install Prompt — for Codex Idle

Codex tidak auto-install skill. Lu perlu kasih prompt install dulu satu kali
biar Codex bikin folder `.codex/skills/aiproxy-autonomous-loop/` di repo.

Setelah itu baru pakai trigger prompt dari `PROMPT_TEMPLATES.md`.

---

## STEP 1 — Install Prompt (jalankan SEKALI di repo AIPROXY)

Copy-paste persis ini ke Codex:

```
Setup custom skill di repo ini. Lakukan SEMUA ini dalam satu sesi tanpa nanya
konfirmasi:

1. Buat folder `.codex/skills/aiproxy-autonomous-loop/` di root repo.

2. Tulis file `.codex/skills/aiproxy-autonomous-loop/SKILL.md` dengan isi yang
   gua kasih di message berikutnya (atau ambil dari attachment SKILL.md kalau
   sudah gua upload).

3. Tulis file `.codex/skills/aiproxy-autonomous-loop/PROMPT_TEMPLATES.md`.

4. Tulis file `.codex/skills/aiproxy-autonomous-loop/INSTALL.md`.

5. Tambahkan baris berikut ke `.gitignore` KALAU belum ada
   (jangan duplikat) — JANGAN ignore folder skill itu sendiri:
       # nothing to add — skill folder is intentionally tracked

6. Verifikasi struktur:
       ls -la .codex/skills/aiproxy-autonomous-loop/
   harus muncul 3 file: SKILL.md, PROMPT_TEMPLATES.md, INSTALL.md

7. Commit ke branch saat ini dengan pesan:
       chore(skill): add aiproxy-autonomous-loop custom skill for Codex idle

8. Konfirmasi balik dalam 1 baris saja:
       [install ok] aiproxy-autonomous-loop ready — gunakan trigger phrase
       untuk eksekusi.

JANGAN edit file lain. JANGAN jalankan task apapun dari folder tasks/.
JANGAN aktifkan skill ini sekarang. Cukup install saja.
```

Setelah Codex selesai install, lanjut ke STEP 2 di bawah.

---

## STEP 2 — Run Prompt (idle / tinggal pergi)

Setelah install OK, kapan saja lu mau lepas Codex jalan sendiri, pakai prompt
ini:

```
Aktifkan skill aiproxy-autonomous-loop dari `.codex/skills/aiproxy-autonomous-loop/SKILL.md`.

Mode: UNATTENDED — gua mau tinggal pergi.

Eksekusi semua plan yang sudah kamu buat sampai selesai dan lakukan testing
sampai tidak ada error apapun. JANGAN PERNAH BERHENTI kalau semua tasks dari
plan belum diproses.

Aturan:
- Kalau satu task gagal verify, retry sampai 5x. Tetap gagal? `git restore`,
  mark `[!]` di TASK_STATUS.md, LANJUT ke task berikutnya. Loop JANGAN
  berhenti.
- Working tree harus bersih sebelum task berikutnya.
- One task = one commit untuk yang sukses; one commit terpisah untuk update
  TASK_STATUS.md.
- Setelah semua task diproses (entah `[x]` atau `[!]`), tulis LOOP_REPORT.md
  di root repo dengan rincian apa yang sukses & apa yang blocked.

Ikuti AGENT_RULES.md, BUILD_RULES.md, REF_IMPLEMENTATION_RULES.md, dan
PROJECT_CONSTRAINTS.md secara strict. JANGAN nanya konfirmasi apapun selama
loop.

Mulai sekarang.
```

---

## STEP 3 — Saat Lu Balik

Cek file `LOOP_REPORT.md` di root repo. Format-nya:

```
## Summary
- Total tasks attempted: 17
- ✓ Completed: 14
- ✗ Blocked: 3

## Blocked (require human review)
### T2.7 — <title>
- First root error: ...
- File: ...
- Attempts: 5
```

Untuk task yang `[!]` blocked, lu review manual, fix, lalu re-run STEP 2.
Skill akan otomatis skip yang sudah `[x]` dan lanjut yang masih `[ ]`.

Kalau mau retry yang blocked juga:

```
Aktifkan skill aiproxy-autonomous-loop. Retry semua task `[!]` di phase ini.
```

---

## Troubleshooting

| Gejala                                          | Solusi                                       |
|--------------------------------------------------|-----------------------------------------------|
| Codex masih nanya konfirmasi tiap task          | Cek SKILL.md ke-load? Cek output line pertama harus `[aiproxy-autonomous-loop] active` |
| Loop berhenti di task pertama yang fail          | Skill versi lama — re-install pakai SKILL.md terbaru |
| Working tree berantakan setelah loop             | Skill versi lama — pastikan ada §2.2 Step A pre-flight cleanup |
| LOOP_REPORT.md tidak dibuat                     | Skill versi lama — update SKILL.md           |

— END —
