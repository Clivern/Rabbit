# Demo setup against a (new) backend

Runbook for pointing the demo at a Sloth backend. You need a **user API key**
(sent as `X-API-Key`) and the backend base URL.

## Order of operations

1. Find the workspace id.
2. Create the access key.
3. Update `run.local.sh` (URL, workspace, access key, agent secret).
4. Run the demo — **it registers all agents, checks, instances and
   intentions on boot, so never pre-create agents via the API.**
   The assistant also issues its own `default` api_key identity from
   `SLOTH_AGENT_SECRET` on boot.

## 1. Find the workspace

```bash
curl -s -m 30 -H "X-API-Key: <user-api-key>" \
  $SLOTH_URL/api/v1/workspaces
```

Pick the workspace `id` (e.g. `"role": "owner"`).

## 2. Create the access key

The demo (`util.py`) calls agent upsert, protection, checks, instance
register/heartbeat and intention list/create, so the key needs exactly:

- `CAN_UPSERT_AGENT`
- `CAN_REGISTER_INSTANCE`
- `CAN_REPORT_CHECK`
- `CAN_LIST_INTENTIONS`
- `CAN_CREATE_INTENTION`

```bash
curl -s -m 30 -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: <user-api-key>" \
  -d '{"name":"demo","permissions":["CAN_UPSERT_AGENT","CAN_REGISTER_INSTANCE","CAN_REPORT_CHECK","CAN_LIST_INTENTIONS","CAN_CREATE_INTENTION"]}' \
  $SLOTH_URL/api/v1/workspaces/<workspace-id>/keys
```

Copy the `key` from the response (shown only once).

## 3. Update `run.local.sh`

```bash
export SLOTH_URL=<backend-url>
export SLOTH_WORKSPACE="<workspace-id>"
export SLOTH_ACCESS_KEY="<access-key>"
export SLOTH_AGENT_SECRET="<assistant-secret>"  # any random value, min 8 characters
export MODEL_API_KEY="<model-key>"
```

`MODEL_API_URL` picks the model endpoint: the Meta responses API when the URL
ends with `/responses` (default `https://api.meta.ai/v1/responses`), OpenAI
chat completions otherwise. To run the assistant on OpenRouter instead:

```bash
export MODEL_API_URL="https://openrouter.ai/api/v1/chat/completions"
export MODEL_API_KEY="<openrouter-key>"
export ASSISTANT_MODEL="openai/gpt-4o-mini"
```

## 4. Run

```bash
./demo/run.sh
```

This registers `assistant`, `weather`, `stock`, `users`, their checks and
instances, plus the assistant's allow-intentions. The assistant also
issues its own `default` api_key identity from `SLOTH_AGENT_SECRET`
(`CAN_UPSERT_AGENT` already covers issuing), so restarts just work —
no dashboard or curl step.

If the chat says Sloth doesn't recognize the assistant's secret, the
stored secret no longer matches `SLOTH_AGENT_SECRET`: revoke the
assistant's identities under Agents > assistant > Identities in the
dashboard, then restart `./demo/run.sh` so the identity is re-issued
from the env value.
