# Product Requirements Document (PRD)  
**Feature: External Project Integration (GitHub/GitLab/Jira) & Issue Linking in Time Entries**

---

## 1. Overview
Timeasy aims to position itself as a time tracking tool tailored for developers.  
To achieve this, each **Timeasy Project** can be linked to exactly one external project or repository from **GitHub, GitLab, or Jira**.  

When users enter a description for a time entry:  
- They will receive autocomplete suggestions from **past time entry descriptions** within the project.  
- If the input matches an **Issue/Key pattern** (via Regex), the system will automatically **resolve the issue** from the linked external project/repository and link it to the time entry.  

If no match is found, the entry remains as free text.  

---

## 2. Goals
- Provide **developers** with seamless integration to their daily tools (GitHub/GitLab/Jira).  
- Enable **issue linking** without requiring manual search.  
- Ensure the solution works both in the **WebApp (Angular)** and the **Mobile App (Flutter)** with minimal overhead.  
- Keep the system **performant and robust**, even for projects with many issues.

---

## 3. Scope

### 3.1 Project Linking
- Each **Timeasy Project** can be linked to **exactly one external project/repo**.  
- Supported providers: **GitHub, GitLab, Jira**.  
- Configuration is done in the **WebApp** via the **Project Edit screen**.  
- Users authenticate via **OAuth** (stored securely server-side).  
- Once connected, Timeasy synchronizes metadata of the external issues.  
- **Implementation Note:** Code must access the **actual APIs** of GitHub, GitLab, and Jira to fetch issue data.  

### 3.2 Description Autocomplete
- **Default Autocomplete**: Suggestions from past time entry descriptions of the same project.  
  - Angular: fetch from backend (`/descriptions/suggest`).  
  - Flutter: local prefix search (already implemented).  
- **Issue Recognition**:  
  - Regex-based detection of issue keys/numbers:
    - GitHub/GitLab: `#123` or `owner/repo#123`  
    - Jira: `ABC-123`  
  - On match → resolve issue via backend (`/issues/resolve`).  
  - If issue exists → show chip “#123 · Fix login bug”.  
  - If unresolved → show “#123 (pending)”.  

### 3.3 Time Entry Linking
- If an issue is recognized, the **time entry** is saved with `external_issue_id`.  
- If not, the description is saved as **plain text**.  
- Flutter offline flow:
  - Save with `pendingExternalRef` if unresolved.  
  - At next sync, backend resolves pending references in batch.

---

## 4. User Stories

### 4.1 Project Linking
- As a user, I can link a Timeasy Project to **one GitHub/GitLab/Jira project**.  
- As a user, I can authenticate via OAuth to allow Timeasy to access issues.  
- As a user, I can change or remove the external link at any time.  

### 4.2 Autocomplete & Issue Linking
- As a user, I see **suggestions from past descriptions** when entering text.  
- As a user, if I type `#123` or `ABC-123`, Timeasy shows the corresponding issue.  
- As a user, I can confirm or remove the issue link.  
- As a user, if offline, I still see suggestions from local history.  

### 4.3 Sync & Pending Resolution
- As a user, if I type `#123` offline, the app stores it and links it later when synced.  
- As a user, I am informed if an issue reference could not be resolved.  

---

## 5. Technical Requirements

### 5.1 Data Model
**ExternalConnection**
- `id`
- `project_id` (FK → Timeasy Project)  
- `provider` (`github`, `gitlab`, `jira`)  
- `project_ref` (e.g., `owner/repo`, `projectKey`)  
- `oauth_token` (encrypted)

**ExternalIssue**
- `id`  
- `project_id`  
- `provider`  
- `key_or_number` (`#123`, `ABC-123`) → **UNIQUE (project_id, provider, key_or_number)**  
- `title`  
- `state` (open/closed/etc.)  
- `updated_at`  

**TimeEntry**
- `id`  
- `project_id`  
- `description`  
- `external_issue_id` (nullable)  

---

### 5.2 Backend APIs

#### 1) Description Suggestions
```http
GET /api/projects/{pid}/descriptions/suggest?q=deb&limit=10
Response: ["debug oauth flow", "deploy staging"]
```
- Returns distinct project descriptions filtered by prefix.
- Sorted by frequency + recency.

#### 2) Issue Resolve
```http
GET /api/projects/{pid}/issues/resolve?input=#123
Response: { "issueId":"gh:123", "key":"#123", "title":"Fix login bug", "state":"open" }
```
- Lookup order:  
  1. Local DB (`ExternalIssue`)  
  2. Live provider API call (GitHub/GitLab/Jira) if not cached  
  3. If unresolved → return 202 `{"key":"#123", "status":"pending"}`  

---

### 5.3 Regex Patterns
- **GitHub/GitLab**: `(?<!\w)#(\d+)`  
- **Jira**: `([A-Z][A-Z0-9]+-\d+)`  

---

### 5.4 Angular (WebApp)
- Debounced (150ms) request to `/descriptions/suggest`.  
- Regex detection in text field: on match → `/issues/resolve`.  
- Display chip with issue title if resolved.  

### 5.5 Flutter (Mobile App)
- Local prefix search for past descriptions.  
- Regex detection in Dart.  
- If online → resolve immediately via API.  
- If offline → store `pendingExternalRef`, resolved at next sync.  

---

### 5.6 Sync & Caching
- **Backfill**: Import only “open + last 90 days” issues initially.  
- **Polling**: Refresh issues at configurable intervals (e.g., every 10–15 minutes).  
- **Cache**: LRU + negative cache for failed lookups.  
- **Offline Sync**: Pending refs resolved in batch.  

---

## 6. Edge Cases
- Multiple matches in one description: take first match, allow manual selection later.  
- Connection changed → old pending refs invalidated.  
- Issue closed/renamed → link persists, but metadata (title/state) updates via polling.  

---

## 7. Non-Goals
- Full-text issue search by title (not in MVP).  
- Multi-repo linking per Timeasy Project (always exactly one).  
- Webhook integration (only polling for now).  

---

## 8. Success Metrics
- % of time entries successfully linked to external issues.  
- Latency of autocomplete (<200ms).  
- Sync success rate for pending refs.  
