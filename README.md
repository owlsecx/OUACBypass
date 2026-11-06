# 🦉 OUACBypass

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Linux-0a0c10?style=flat-square"/>
  <img src="https://img.shields.io/badge/Category-Attack%20/%20Offensive-blue?style=flat-square"/>
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square"/>
  <img src="https://img.shields.io/badge/Part%20of-OwlSec%20Toolkit-7b5ea7?style=flat-square"/>
  <img src="https://img.shields.io/badge/Version-1.0-cyan?style=flat-square"/>
</p>

> **OUACBypass** — UAC Bypass Technique Catalog ║

---

> ⚠️ **AUTHORISED USE ONLY** — Use only on systems you own or have explicit written permission to test. Unauthorised use is illegal.

---

## 📌 Overview

OUACBypass — UAC Bypass Technique Catalog. OWL Batch-3 offensive security toolkit. AUTHORIZED TESTING ONLY.

---

## ✨ Features

- Offline catalog of Windows UAC bypass techniques
- Configuration/registry-oriented scanner inputs
- Howto and path-focused technique lookup
- Authorized lab reference for defenders and red teamers
- List/catalog modes for quick triage
- Export technique notes and scan findings

---

## ⚙️ Requirements

- **Linux**
- **Go 1.21+** (stdlib only; `CGO_ENABLED=0`)

---

## 🚀 Usage

```bash
go build -ldflags="-s -w" -o OUACBypass OUACBypass.go
./OUACBypass --help
./OUACBypass --list
./OUACBypass --catalog --format md
./OUACBypass --input export.reg --format json
```

---

## 📦 Part of OwlSec Toolkit

This tool is part of the **OwlSec** suite — a collection of 300+ security and privacy tools.

🔗 [owlsec.org](https://owlsec.org)

---

## ©️ License

MIT License — © Khaled S. Haddad
