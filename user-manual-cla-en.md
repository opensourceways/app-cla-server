# CLA System User Manual

[app-cla-server](https://github.com/opensourceways/app-cla-server) manages Contributor License Agreement (CLA) signing and verification for open source projects, built with Go and the Beego framework.

---

## Roles and Resources Overview

The system operates around four roles with distinct responsibilities:

| Role | Permission Identifier | Responsibilities |
|------|-----------------------|-----------------|
| Community Manager | `owner of org` | Create/delete Links, manage CLA templates, add corp admins |
| Corporation Admin | `corporation administrator` | Sign CCLA on behalf of company, manage employee signings, add employee managers |
| Employee Manager | `employee manager` | Approve/deactivate employee signing records |
| Contributor | — | Sign individual CLA (ICLA) or employee CLA |

Core system resources:

| Resource | Description |
|----------|-------------|
| Link | Configuration unit binding a GitHub/Gitee org or repo to a CLA — the entry point for all signing operations |
| CLA Template | Agreement file (PDF) attached to a Link; two types: `individual` and `corporation` |
| Corp Signing | Record created when a corp admin signs the CCLA on behalf of their company |
| Individual Signing | Record created when an individual contributor signs the ICLA |
| Employee Signing | Record created when an employee signs the employee CLA; requires corp admin approval to activate |

---

## Onboarding Flowchart

```
                              Start
                                │
                                ▼
┌──────────────────────────────────────────────────────────┐
│ Step 1: Determine Signing Type                           │
│ You: Identify contributor type — individual or employee  │
│ Acceptance: Clear on ICLA flow vs CCLA + employee flow   │
└───────────────────────────┬──────────────────────────────┘
                             │
               ┌─────────────┴─────────────┐
               │                           │
               ▼                           ▼
      ┌─────────────────┐       ┌──────────────────────┐
      │  Individual      │       │  Corporate Employee   │
      └────────┬────────┘       └──────────┬───────────┘
               │                           │
               ▼                           ▼
┌──────────────────────────┐  ┌─────────────────────────────┐
│ Step 2A: Sign ICLA        │  │ Step 2B: Corp Admin Signs   │
│ You: Visit signing page,  │  │ CCLA                        │
│     fill in info, submit  │  │ You (corp admin): Visit     │
│     verification code     │  │     signing page, fill in   │
│ Acceptance: Signing record│  │     company info            │
│     created successfully  │  │ Acceptance: Corp signing    │
└─────────────┬────────────┘  │     record created          │
              │                └────────────┬────────────────┘
              ▼                             │
┌──────────────────────────┐               ▼
│ Step 3A: Verify Status    │  ┌─────────────────────────────┐
│ You: Submit PR, system    │  │ Step 3B: Employee Signs CLA  │
│     auto-checks signing   │  │ You (employee): Visit signing│
│ Acceptance: PR passes CLA │  │     page, fill in info,      │
│     check                 │  │     submit verification code │
└──────────────────────────┘  │ Acceptance: Employee signing │
                               │     record pending approval  │
                               └────────────┬────────────────┘
                                            │
                                            ▼
                               ┌─────────────────────────────┐
                               │ Step 4B: Corp Admin Approves │
                               │ You (corp admin): Log in,    │
                               │     activate employee signing│
                               │ Acceptance: Employee signing │
                               │     status becomes active    │
                               └────────────┬────────────────┘
                                            │
                                            ▼
                               ┌─────────────────────────────┐
                               │ Step 5B: Verify Status       │
                               │ You: Submit PR, system       │
                               │     auto-checks signing      │
                               │ Acceptance: PR passes CLA    │
                               │     check                    │
                               └─────────────────────────────┘
```

---

## Community Manager Operations

Community managers initialize the CLA integration. Requires a token with `owner of org` permission.

### Create a Link

A Link is the core configuration unit that binds a code platform org/repo to a CLA agreement.

**What you need to do**:

Send a request to `POST /v1/link/` with the following body:

```json
{
  "platform":     "github",
  "org_id":       "your-org-name",
  "repo_id":      "",
  "org_alias":    "Your Org Display Name",
  "org_logo":     "https://example.com/logo.png",
  "project_url":  "https://github.com/your-org"
}
```

- `platform`: Code platform — `github` or `gitee`
- `org_id`: Organization name on GitHub/Gitee
- `repo_id`: Leave empty for org-level Link; fill in repo name for repo-level Link
- `org_alias`: Organization name shown to signers
- `project_url`: Project homepage URL, included in signing notification emails

**How to verify**:

- Response returns `201`
- `GET /v1/link/` list includes the new Link with its `link_id`

---

### Upload a CLA Template

Each Link requires at least one CLA template (PDF file), one for individual and one for corporation.

**What you need to do**:

Send a `multipart/form-data` request to `POST /v1/cla/:link_id`:

```
link_id:   <link_id from previous step>
apply_to:  individual   (or corporation)
file:      <CLA PDF file>
fields:    <JSON array defining fields signers must fill in>
```

Example `fields` for an individual CLA:

```json
[
  {"id": "name",    "title": "Full Name",   "type": "string", "required": true},
  {"id": "email",   "title": "Email",       "type": "string", "required": true},
  {"id": "company", "title": "Company",     "type": "string", "required": false}
]
```

**How to verify**:

- Response returns `201`
- `GET /v1/cla/:link_id` returns a list including the uploaded template with its `cla_id`

---

### Add a Corporation Admin

After a company completes CCLA signing, the community manager assigns an admin account for that company.

**What you need to do**:

Send a request to `POST /v1/corporation-manager/:link_id/:signing_id` (no body required).

- `signing_id`: Corp signing record ID, retrieved from `GET /v1/corporation-signing/:link_id`

**What we do**:

- System generates a corp admin account (with initial password) and sends a notification email to the corporate contact

**How to verify**:

- Response returns `202`
- Corporate contact receives an email with account credentials and initial password

---

## Corporation Admin Operations

The corporation admin signs the CCLA on behalf of the company and manages employee signing records.

### Sign the Corporate CLA (CCLA)

**What you need to do**:

1. Request a verification code — send to `POST /v1/corporation-signing/:link_id/code`:

```json
{"email": "corp-contact@example.com"}
```

2. Submit the signing — send to `POST /v1/corporation-signing/:link_id/`:

```json
{
  "cla_id":            "<cla_id>",
  "verification_code": "<received code>",
  "email":             "corp-contact@example.com",
  "corporation_name":  "Example Corp Ltd.",
  "name":              "John Smith",
  "address":           "123 Main St, City, Country",
  "date":              "2026-05-13"
}
```

**What we do**:

- Validate the verification code
- Create the corp signing record
- Generate a signed PDF and send it to the corporate contact email

**How to verify**:

- Response returns `201`
- Corporate contact email receives a signing confirmation with PDF attachment

---

### Log In as Corporation Admin

After the community manager adds the corp admin, use the credentials from the notification email to log in.

**What you need to do**:

Send a request to `POST /v1/corporation-manager/`:

```json
{
  "account":  "corp-contact@example.com",
  "password": "<initial password from email>"
}
```

**How to verify**:

- Response returns `200` with `link_id` and account basic info
- `access_token` and `csrf_token` are written to cookies — include both in subsequent requests

---

### Manage Employee Signing Records

After an employee signs the employee CLA, the corp admin must approve and activate the record.

**What you need to do**:

View pending employee list: `GET /v1/employee-signing/`

Activate an employee signing — send to `PUT /v1/employee-signing/:signing_id`:

```json
{"enabled": true}
```

Deactivate an employee signing: set `enabled` to `false`.

**What we do**:

- On activation: system sends "CLA signing activated" notification email to the employee
- On deactivation: system sends "CLA signing deactivated" notification email to the employee

**How to verify**:

- Response returns `202`
- Employee receives a status change notification email
- `GET /v1/employee-signing/` list shows updated status for that employee

---

### Add Employee Managers

The corp admin can promote employees to employee manager role to delegate approval responsibilities.

**What you need to do**:

Send a request to `POST /v1/employee-manager/`:

```json
{
  "managers": [
    {"email": "manager@example.com", "name": "Jane Doe"}
  ]
}
```

**What we do**:

- System sends an authorization notification email to the new employee manager with login credentials

**How to verify**:

- Response returns `201`
- New employee manager receives authorization email
- `GET /v1/employee-manager/` list includes the new manager

---

## Contributor Operations

### Sign Individual CLA (ICLA)

For contributors contributing as individuals, not representing any company.

**What you need to do**:

1. Fetch signing page info (CLA content and field definitions):

```
GET /v1/link/:link_id/individual
```

2. Request a verification code:

```json
POST /v1/individual-signing/:link_id/code
{"email": "contributor@example.com"}
```

3. Submit the signing:

```json
POST /v1/individual-signing/:link_id/
{
  "cla_id":            "<cla_id>",
  "verification_code": "<verification code>",
  "email":             "contributor@example.com",
  "name":              "Alex Chen",
  "date":              "2026-05-13"
}
```

**How to verify**:

- Response returns `201`
- `GET /v1/individual-signing/:link_id?email=contributor@example.com` returns `{"signed": true}`

---

### Sign Employee CLA

For contributors whose company has already signed the CCLA and who are contributing as corporate employees.

**What you need to do**:

1. Confirm your company has completed CCLA signing and get the company's `signing_id`:

```
GET /v1/corporation-signing/:link_id/corps/:corp_email_domain
```

2. Request a verification code:

```json
POST /v1/employee-signing/:link_id/:signing_id/code
{"email": "employee@example-corp.com"}
```

3. Submit the signing:

```json
POST /v1/employee-signing/:link_id/
{
  "cla_id":            "<cla_id>",
  "verification_code": "<verification code>",
  "email":             "employee@example-corp.com",
  "name":              "Bob Wang",
  "date":              "2026-05-13"
}
```

**What we do**:

- System sends a signing confirmation email to the employee
- System sends a "pending approval" notification email to corp admin(s)

**How to verify**:

- Response returns `201`
- After corp admin approves the record, PRs will pass the CLA check

---

## FAQ

**Q1: Signing returns `resigned` error — what does that mean?**

That email has already signed the CLA under this Link. Duplicate signing is not allowed. Contact the community manager if you need to update your signing information.

**Q2: Verification code expired or incorrect — what should I do?**

Verification codes have a short validity window (typically 10 minutes). Call the `/code` endpoint again to get a new code, then resubmit.

**Q3: Employee signed the CLA but PR still fails the CLA check?**

Employee signing records must be activated by the corp admin (`enabled: true`) before they take effect. Confirm the corp admin has approved the record.

**Q4: Corp admin forgot their password?**

Call `POST /v1/password-retrieval/` to reset the password via the registered email address.

**Q5: How do I check whether an email has signed the ICLA?**

```
GET /v1/individual-signing/:link_id?email=<email>
```

Returns `{"signed": true}` if the email has signed.

**Q6: Corp signing PDF was not received — what to do?**

The community manager can call `PUT /v1/corporation-signing/:link_id/:signing_id` to re-trigger PDF generation and email delivery.

**Q7: Where do I get the `link_id`?**

Community managers call `GET /v1/link/` after logging in to retrieve all Links and their `link_id` values.

---

## Feedback & Support

Submit issues on GitHub:

[https://github.com/opensourceways/app-cla-server/issues](https://github.com/opensourceways/app-cla-server/issues)

Include the following when reporting:
- Your role (community manager / corp admin / contributor)
- The API endpoint and request body you called
- The error code and error message returned
- The `link_id` involved (if applicable)
