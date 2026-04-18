<img alt="OpenGraph for BhariyaAuth" src="Frontend/public/open-graph.jpg"/>

# BhariyaAuth

### A Unified, High-Performance Identity Provider for All Your Services

---

BhariyaAuth is a **modern, self-hostable Identity Provider (IdP)** built to simplify authentication across all your platforms — websites, apps, APIs — under a single, seamless system.

It is designed with a clear goal:

> **Make authentication fast, secure, user-friendly, and inexpensive to run.**

---

## Why BhariyaAuth Exists

Most authentication systems today create friction:

* You log in once… and still need to log in again elsewhere
* Emails are confusing or unreliable
* Infrastructure is expensive and complex
* Each app needs its own auth system

### BhariyaAuth solves this.

✔ One login → all services
✔ Reliable email + retry systems
✔ Lightweight → low hosting cost
✔ One auth server → everything

---

## What You Get

### 🔐 Seamless Authentication

* One account across all your services
* Works across websites, apps, and APIs
* No repeated logins

---

### ⚡ Fast & Efficient

* Built in Go using Fiber
* Minimal overhead
* Designed for **low latency + high throughput**

---

### 🔁 Smart Retry System

* Automatic retries for:

    * OTP delivery
    * email sending
    * temporary failures
* Users don’t get stuck due to transient issues

---

### 📧 Clean Email Experience

* Clear, human-readable emails
* Login alerts, OTPs, resets
* Context-aware messaging (device, IP, etc.)

---

### 🧠 Intelligent Rate Limiting

* Not just request-based — **weight-based**
* Prevents abuse without hurting real users

---

### 🔄 Session Control

* Device-based session tracking
* Revoke specific sessions
* Revoke all sessions instantly

---

### 🔐 Strong Security

* AES-GCM encrypted tokens (not JWT)
* CSRF protection
* OTP verification
* Token isolation by purpose

---

## How It Works

```text
Client (Web/App)
        ↓
   BhariyaAuth API
        ↓
PostgreSQL (users, devices)
        ↓
Redis (optional support)
        ↓
In-memory systems:
   - OTP store
   - Rate limiter
```

---

## Authentication Flows (Simple & Clear)

### 🆕 Sign Up

```text
Step 1 → Enter name, email, password  
Step 2 → Verify OTP  
→ Account created
```

---

### 🔐 Sign In

```text
Step 1 → Choose login method:
   • OTP
   • Password

Step 2 → Verify  
→ Logged in
```

---

### 🔁 Password Reset

```text
Step 1 → Request reset  
Step 2 → OTP + new password  
→ Password updated
```

---

### 🔐 Multi-Factor Authentication (MFA)

```text
Step 1 → Request OTP  
Step 2 → Verify  
→ Access granted
```

---

### 📱 Session Management

* View all devices
* Revoke specific session
* Revoke all sessions

---

## Opaque Token System (Better than JWT)

BhariyaAuth uses **encrypted structured tokens** instead of JWT.

### Why this matters:

* 🔒 Fully encrypted (not just signed)
* 🧩 Typed tokens (no misuse)
* ⚡ Stateless and fast

### Token Types:

* Access Token (short-lived)
* Refresh Token (rotating)
* MFA Token
* SSO Token
* Sign-in / Sign-up Tokens

---

## For End Users

When a service uses BhariyaAuth:

### You’ll experience:

* Login once → access everything
* Choose OTP or password
* Get notified on new logins
* Manage sessions easily

### No more:

* repeated logins
* confusing emails
* blocked actions due to rate limits

---

## For Developers & Hosters

### Why BhariyaAuth is powerful:

#### 🧩 One Auth Server → All Services

No need to duplicate auth logic across apps.

---

#### 💸 Lower Hosting Costs

* Minimal dependencies
* Efficient memory usage
* No forced horizontal scaling

---

#### ⚙️ No External Complexity

* Built-in:

    * rate limiter
    * OTP system
    * retry logic

---

#### 🧼 Clean Codebase

* Modular structure
* Easy to maintain
* Designed for extensibility

---

## Getting Started

---

### 🔧 Requirements

* Go (latest)
* PostgreSQL
* Redis (optional)
* Linux (recommended for UNIX sockets)

---

### 📥 Clone the Project

```bash
git clone https://github.com/BhaskarPanja93/BhariyaAuth.git
cd BhariyaAuth
```

---

### 🔐 Configure Environment

Create your secrets:

```
SQLUser=...
SQLPassword=...
SQLDBName=...
SQLHost=...
SQLPort=...

RedisHost=...
RedisPort=...

AESGCMKey=your_32_byte_secret_key
```

---

### Run the Server

#### Option 1: Default (Port 3000)

```bash
go run main.go
```

#### Option 2: UNIX Socket

```bash
go run main.go -bind /tmp/bhariya.sock
```

---

### 🌍 API Base Path

```text
/auth/api/
```

---

## 📡 API Modules

| Module            | Purpose            |
| ----------------- | ------------------ |
| `/signup`         | Register users     |
| `/signin`         | Login              |
| `/password-reset` | Reset password     |
| `/session`        | Manage sessions    |
| `/mfa`            | Multi-factor auth  |
| `/sso`            | SSO authentication |

---

## Security Highlights

* AES-GCM encryption (confidential + integrity)
* CSRF protection (double-submit)
* Token-type enforcement
* Device-based sessions
* OTP expiration + retry limits

---

## Performance

* Low memory footprint
* Efficient DB pooling
* Minimal latency
* Suitable for:

    * small → medium scale without horizontal scaling

---

## Updates & Roadmap

BhariyaAuth is actively evolving.

### Upcoming:

* Per device instead of per IP rate limiting
* Admin dashboard
* Better monitoring

---

## Contributing

Contributions are welcome but aren't guaranteed to be implemented.

You can help with:

* security improvements
* performance optimizations
* documentation
* integrations


---

## Final Thought

BhariyaAuth is built to be:

> **Simple to run. Efficient to scale. Pleasant to use.**

A single authentication layer for everything you build.

---
