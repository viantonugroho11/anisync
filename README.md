# anisync 🌿

**anisync** is a lightweight distributed synchronization library for Go,
built on top of Redis.

It provides:
- Distributed locks
- Auto-renewal (heartbeat)
- Leader election
- Context-aware acquisition
- Prometheus metrics

Designed for **cron jobs**, **workers**, and **high-availability services**.

---

## ✨ Features

- 🔒 Distributed lock (safe release via Lua)
- 🔁 Auto-renew / heartbeat
- ⏳ Blocking & non-blocking acquire
- 👑 Leader election
- 📊 Prometheus metrics
- 🧪 Fully unit-tested (miniredis)

---

## 📦 Installation

```bash
go get github.com/yourorg/anisync
