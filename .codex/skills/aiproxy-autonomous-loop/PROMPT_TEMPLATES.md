# Prompt Templates — Idle / Unattended Mode

Skill ini dirancang untuk **TINGGAL PERGI**. Loop tidak akan pernah berhenti
karena satu task gagal — task yang gagal di-revert, ditandai `[!]`, loop lanjut.

Flow lengkap:

1. **SEKALI** — install skill ke repo via prompt di `INSTALL_PROMPT.md`.
2. **TIAP MAU JALAN** — pakai salah satu prompt di bawah.
3. **PAS BALIK** — cek `LOOP_REPORT.md` di root repo.

---

## 1. ⭐ Full Unattended Loop (rekomendasi utama)

```
Aktifkan skill aiproxy-autonomous-loop dari `.codex/skills/aiproxy-autonomous-loop/SKILL.md`.

Mode: UNATTENDED — gua tinggal pergi.

Eksekusi semua plan sampai selesai. JANGAN PERNAH BERHENTI kalau semua tasks
belum diproses.

- Task gagal verify? Retry 5x. Masih gagal? `git restore`, mark `[!]`, lanjut
  task berikutnya.
- Working tree harus clean sebelum tiap task.
- One task = one commit.
- Setelah semua task diproses, tulis LOOP_REPORT.md di root repo.

Ikuti AGENT_RULES.md, BUILD_RULES.md, REF_IMPLEMENTATION_RULES.md, dan
PROJECT_CONSTRAINTS.md strict. JANGAN nanya konfirmasi apapun.

Mulai sekarang.
```

## 2. Resume (lanjut dari task yang masih `[ ]`)

```
Aktifkan skill aiproxy-autonomous-loop. Mode: UNATTENDED.
Resume dari task pertama yang masih `[ ]` atau `[~]` di TASK_STATUS.md untuk
phase aktif. Skip yang `[x]` dan `[!]`. Loop sampai habis. Tulis LOOP_REPORT.md
di akhir.
```

## 3. Retry Blocker (kalau lu udah fix root cause)

```
Aktifkan skill aiproxy-autonomous-loop. Mode: UNATTENDED.
Retry semua task `[!]` di phase aktif. Reset status mereka ke `[ ]` dulu,
lalu jalankan loop normal. Tulis LOOP_REPORT.md di akhir.
```

## 4. Batch Spesifik (kalau lu mau scope tertentu aja)

```
Aktifkan skill aiproxy-autonomous-loop. Mode: UNATTENDED.
Eksekusi HANYA T2.1 sampai T2.5 berurutan. Setiap task: verify, commit kalau
green, mark `[!]` kalau gagal setelah 5 retry. Loop sampai T2.5 selesai
diproses. Tulis LOOP_REPORT.md.
```

## 5. Full Phase + Open PR otomatis

```
Aktifkan skill aiproxy-autonomous-loop. Mode: UNATTENDED.
Eksekusi seluruh Phase aktif sampai semua task diproses. Setelah loop selesai
dan LOOP_REPORT.md ditulis:
- Kalau 100% `[x]` (zero blocked), push branch dan open PR ke main dengan
  title `Phase <N> — <phase title>` dan body = isi LOOP_REPORT.md.
- Kalau ada `[!]`, JANGAN open PR. Push branch saja dan stop. Lu review dulu.
```

---

## ❌ JANGAN dipakai (kepicu Codex built-in planning)

- ❌ `/plan ...`
- ❌ `enter plan mode`
- ❌ `think step by step before coding`
- ❌ `propose a plan first`
- ❌ `ask me before each task`
- ❌ `confirm before continuing`

Skill ini sudah handle planning + execution loop sekaligus. Cukup trigger.
