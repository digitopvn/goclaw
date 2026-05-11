# Upstream `nextlevelbuilder/goclaw` — Important Issues & PRs

Source: https://github.com/nextlevelbuilder/goclaw
Scanned: 2026-05-11 (latest 100 issues + 100 PRs)
Purpose: Triage list for porting into `digitopvn/goclaw`.

Legend: 🔴 critical · 🟠 high · 🟡 medium · 🟢 nice-to-have

---

## 🔴 Security (open issues — must triage first)

| # | Title |
|---|-------|
| 1134 | [Security] Unauthenticated HTTP Admin Access and Forged Webhook Execution When Secrets Are Left Unconfigured |
| 1133 | [Security] Team task completion authorization bypass — member completes another member's in-progress task |
| 1132 | [Security] SSRF via Unvalidated TTS Provider API Base Configuration |
| 1121 | [Security] Arbitrary Command Execution inside Sandbox via FsBridge Command Injection |
| 1120 | [Security] Evolution Suggestion Authorization Bypass — Viewer mutates admin-only suggestion state (disables tools) |
| 1119 | [Security] Agent Config Bypass via V3-Flags — Viewer mutates admin-only agent config |
| 1118 | [Security] RoleAdmin Gateway Auth Bypass — Viewer modifies admin-only TTS and Storage config |

---

## 🟠 Critical Bugs (open)

| # | Title |
|---|-------|
| 1136 | Session auto-compaction can fail to prevent context overflow — add overflow-triggered compact+retry + manual compact |
| 1126 | MCP bridge drops SenderID → cron / write_file unusable in Telegram groups (Claude CLI provider) |
| 1131 | Duplicate `write_file` Result — Take Action |
| 1122 | Dashboard refuses to restore backups |
| 1077 | Xoá team thất bại — vault_docs_team_null_scope_fix conflicts with CHECK constraint vault_documents_scope_consistency |
| 1076 | Sao lưu tenant thất bại — config_secrets missing `id`, hooks + tenant_hook_budget missing from backup registry |
| 1075 | Browser Pairing stuck on `/setup` — RoleOperator can't access GET /v1/providers |
| 1073 | Workspace isolation inverts user/channel hierarchy — artifacts split across channels for same user |
| 1070 | MCP test succeeds but save fails for local/private MCP URLs due to SSRF validation mismatch |
| 1064 | WhatsApp: multiple channel_instances share one whatsmeow_device row |
| 1058 | Telegram channel permanently broken after proxy timeout on startup — no reconnect attempt |
| 1054 | Master system cannot edit predefined context files for newly-created agent |
| 1047 | DeepSeek V4 Pro: 400 due to missing reasoning_content on Zalo channel (OpenAI-compatible) |
| 1042 | OpenRouter+Anthropic: prompt caching disabled, ~3–4× cost overhead |
| 1034 | After upgrading v3.11.1 from v3.9.2, provider is broken |
| 1029 | Sandbox container name invalid — `@` in WhatsApp chat ID breaks Docker naming |
| 1027 | MCP arg validation false positive — substring match on `-c` blocks legitimate package names |
| 1026 | Flaky MCP server health-check loop tanks platform health, blocks dashboard |
| 994  | `restrict_to_workspace=false` ignored by write_file/read_file/edit |
| 993  | MCP tool routing divergence: deferred MCP tools fail "unknown tool" 1ms on direct invoke |
| 978  | Editing provider Advanced settings (API key change) triggers infinite 500/404 loop — requires DB fix |
| 977  | [Lite v3.9.0] Every chat blocked — "hook blocked user_prompt_submit" due to missing hooks tables |
| 957  | `goclaw doctor` misses managed runtime mismatches that break a seemingly healthy install |
| 956  | Skill dependency checks/install break on macOS — wrong Python/pip runtime resolved |

---

## 🟡 Notable Feature Requests (open)

| # | Title |
|---|-------|
| 1102 | Background Exec with Output Delivery to LLM |
| 1097 | Built-in `wait` tool with millisecond delay parameter |
| 1095 | Pre-Auth Webhook Interceptor for Channels (Prevent LLM Denial-of-Wallet) |
| 1087 | Detect & auto-activate skills via slash-command syntax in user prompts |
| 1086 | Chain LLM providers with automatic fallback when primary fails |
| 1036 | Configurable message debounce timer for continuous chat experience |
| 970  | Support for OpenAI Responses API |
| 966  | Support Zalo OA for non-bot (phone-number) accounts |
| 930  | Empty LLM Response Retry at Pipeline Level |
| 918  | Add custom provider_type for generic OpenAI-compatible endpoints |
| 907  | Integrate `microsoft/markitdown` as lightweight document reader |
| 900  | Updating "Runtime & Packages" flow (partially shipped in this repo) |
| 890  | Integrate Google Magika for AI-powered file type detection / malware defense |
| 883  | Simplify GitHub Integration for Agents |
| 880  | MCP Bundle (.mcpb) Import Support — one-click MCP server install |
| 879  | MCP OAuth Integration — OAuth 2.1 via existing ChatGPT/OpenAI subscriptions |
| 1112 | Add Portuguese (Brazil) pt-BR localization |
| 1098 | `GOCLAW_CRON_JOB_TIMEOUT` env var for configurable cron timeout |

---

## 🟠 Important Open PRs (candidates to port)

### Security / correctness fixes
| # | Title |
|---|-------|
| 1140 | fix(exec): exempt venv python interpreter from `.goclaw/` path deny |
| 1139 | fix(cron): invert stateless gate to match UI label |
| 1129 | fix(cron): stamp + replay creator sender/role across cron fires |
| 1128 | fix(announce): propagate sender + role through team-task announce re-ingress |
| 1115 | fix(mcp): exact-match dangerous flags, stop `-c` substring false positives |
| 1111 | fix(gateway): re-apply tool rate limiter after system_configs overlay |
| 1105 | fix(perms): propagate sender/role through MCP bridge & delegate (#915) |
| 1094 | fix(acp): propagate agent_key through MCP bridge so tools resolve identity |
| 1083 | fix(cli): add missing `X-GoClaw-User-Id` header to gateway client |
| 1082 | fix(ops): bot token masking + episodic timeout 120s + path_escape log levels |
| 1079 | fix(http,ws): RBAC for providers read + master predefined context files |
| 1065 | fix(whatsapp): scope whatsmeow_device per channel instance |
| 1063 | fix(providers): resolve HTTP 400 DeepSeek reasoning passback during tool calls |
| 1032 | fix(cron): always reset session before cron runs — stateless flag inverted |
| 1031 | fix(sandbox): sanitize `@` in container names for WhatsApp LID chat IDs |
| 983  | fix(memory): truncate embeddings locally |

### Performance
| # | Title |
|---|-------|
| 1123 | fix(providers): increase HTTP connection pool for embedding provider |
| 1127 | feat(providers): explicit prompt cache for DashScope/Qwen |
| 1044 | perf(providers): cache conversation history on Anthropic requests |

### Features
| # | Title |
|---|-------|
| 1138 | feat(buildah): self-building OCI runtime |
| 1117 | feat(ui): add Compact button to session detail page |
| 1114 / 1099 | feat(config): `GOCLAW_CRON_JOB_TIMEOUT` env override |
| 1113 | feat(provider): add Qiniu provider support |
| 1110 | feat(pkg/tool): `GlobalToolFactoryRegistry` for external tool registration |
| 1109 | feat(channels): add Max Messenger channel |
| 1108 | feat(web_search): `provider` arg to force a specific engine |
| 1106 | feat(telegram): `/addcron /removecron /croners` group cron-permission commands |
| 1100 | feat(channels): Instagram channel integration |
| 1084 | feat(i18n): cascade locale resolution for cron + system-triggered runs |
| 1081 | feat(i18n): Korean (ko) localization |
| 1068 | feat(i18n): full Russian localization |
| 1072 | feat(acp): full Gemini CLI integration — tool exposure + spec compliance + watchdog + GEMINI.md |
| 1066 | feat(auth): inherit `tenant_users.role` for paired sessions |
| 1061 / 1060 / 1059 | feat(channels): Bitrix24 messenger (3-part split) |
| 1048 | feat(channels/zalo): OA OAuth + webhook transport (#966) |
| 1045 | feat(hooks): surface script reason in synthetic block messages |
| 1033 | feat(exec): detect credentialed CLIs in shell chains + allow_chain_exec injection mode |
| 1024 | feat(discord): real-time voice-channel transcription |
| 1016 | feat(host-actions): execline-based dispatcher with hardening |
| 1013 | feat(channels): configurable table rendering for Telegram and WhatsApp |
| 1012 | feat: configurable MCP stdio command allowlist |
| 1011 | feat(skills): privacy/visibility controls for agent-owned skills |
| 989  | feat(pipeline): add tool call authorization mechanism |
| 988  | feat(tools): `send_discord_embed` for rich embeds |
| 987  | feat(channels): `suppress_placeholder` toggle for Discord and Slack |
| 981  | feat(webhooks): HTTP webhooks trigger agents with HMAC + durable callbacks (already merged here) |
| 1028 | Add Lightpanda as alternative to Chrome for browser automation |

### Skills / Windows / DX
| # | Title |
|---|-------|
| 1104 | fix(skills): instruct agent to use exact absolute path for skill files |
| 1103 | fix(skills): normalize backslash in ZIP entry names for Windows |
| 1116 | fix(agent): suppress empty final content instead of literal `"..."` |
| 1101 | fix(agent): add Thinking field to synthetic post-summary assistant message |

---

## Already shipped to `digitopvn/goclaw` (skip)

Per recent commits on `dev`:
- #900 Phase 2a/2b — packages update flow (pip/npm/apk + pkg-helper v2)
- #4 Remote Workstation Runtime (SSH exec + security + audit) ← from upstream #941
- #2 HTTP Webhooks ← from upstream #981
- #3 Packages & CLI Credentials unification

---

## Unresolved questions

- Are the 7 open Security issues (#1118–1121, #1132–1134) already mitigated in `digitopvn/goclaw`'s fork, or do they need targeted patches?
- Locale PRs (#1068 ru, #1081 ko, #1112 pt-BR) — which locales does digitop want to maintain?
- Channels (Instagram #1100, Bitrix24 #1059–1061, Max #1109) — in scope or out of scope for this fork?
- Provider PRs (#1113 Qiniu, #1093 New providers) — desired in fork?
