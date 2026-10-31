"""Register a demo agent with Sloth, keep its heartbeat check passing and let Sloth probe /healthz.

Set SLOTH_URL, SLOTH_WORKSPACE and SLOTH_ACCESS_KEY to register. Without
SLOTH_URL the agent only serves its card and message:send locally.
Set INSTANCE to override the default {name}-{port} instance id, e.g.
weather01.ams01.xai.com.
"""

import asyncio
import os
import secrets
from contextlib import asynccontextmanager

import httpx
from fastapi import FastAPI, HTTPException, Request

URL = os.environ.get("SLOTH_URL", "")
WORKSPACE = os.environ.get("SLOTH_WORKSPACE", "")
ACCESS_KEY = os.environ.get("SLOTH_ACCESS_KEY", "")
ADDRESS = os.environ.get("AGENT_ADDRESS", "127.0.0.1")
INSTANCE = os.environ.get("INSTANCE", "")
TTL = 30


async def heartbeat(client, base, instance, port):
    # Sloth fetches the card when the instance registers, so wait until the agent is serving it.
    while True:
        try:
            await client.get(f"http://127.0.0.1:{port}/.well-known/agent-card.json")
            break
        except httpx.TransportError:
            await asyncio.sleep(0.2)

    await client.put(f"{base}/instances/{instance}", json={"address": ADDRESS, "port": port, "path": "/a2a/v1"})

    while True:
        await client.post(f"{base}/instances/{instance}/checks/heartbeat/pass")
        await asyncio.sleep(TTL / 3)


async def allow(client, name, calls):
    """Allow name to call each agent in calls, skipping pairs that already have an intention."""
    response = await client.get(f"{URL}/api/v1/workspaces/{WORKSPACE}/intentions", params={"limit": 100})
    existing = {item["destination"] for item in response.json()["intentions"] if item["source"] == name}

    for destination in calls:
        if destination in existing:
            continue

        await client.put(
            f"{URL}/api/v1/workspaces/{WORKSPACE}/intentions",
            json={"source": {"agent": name}, "destination": {"agent": destination}, "action": "allow"},
        )


def lifespan(name, version, port, protection=None, calls=None, secret=""):
    """Register the agent on boot; protection, when set, is the credential Sloth must send the agent,
    calls lists the agents it is allowed to call, and secret, when set, is issued as its gateway identity."""

    @asynccontextmanager
    async def run(app):
        if not URL:
            yield
            return

        base = f"{URL}/api/v1/workspaces/{WORKSPACE}/agents/{name}"
        instance = INSTANCE or f"{name}-{port}"

        async with httpx.AsyncClient(headers={"X-Access-Key": ACCESS_KEY}) as client:
            await client.put(base, json={"version": version, "cardPath": "/.well-known/agent-card.json"})

            # The agent issues its own gateway identity with its secret; 409 means one already exists.
            if secret:
                await client.post(f"{base}/identities", json={"name": "default", "type": "api_key", "secret": secret})

            if protection:
                await client.put(f"{base}/protection", json=protection)

            await client.put(f"{base}/checks/heartbeat", json={"type": "ttl", "ttl": TTL})
            await client.put(f"{base}/checks/health", json={"type": "http", "path": "/healthz", "interval": 10})

            if calls:
                await allow(client, name, calls)

            task = asyncio.create_task(heartbeat(client, base, instance, port))
            yield
            task.cancel()

            # Kept registered so a stopped agent goes critical and alerts can be tested.
            await client.delete(f"{base}/instances/{instance}")

    return run


def agent_app(name, display, description, version, port, skills, protection=None):
    """Serve /healthz, the agent card and A2A dispatch for JSON skills.

    Each skill carries id, name, description, tags, examples, func and an
    optional inputModes; the first skill is the default. protection, when
    set, registers with Sloth and is the header every call must send.
    """
    app = FastAPI(lifespan=lifespan(name, version, port, protection=protection))
    funcs = {skill["id"]: skill["func"] for skill in skills}

    @app.get("/healthz")
    def healthz():
        return {"status": "ok"}

    @app.get("/.well-known/agent-card.json")
    def agent_card():
        fields = ("id", "name", "description", "tags", "inputModes", "examples")

        return {
            "name": display,
            "description": description,
            "version": version,
            "supportedInterfaces": [
                {
                    "url": f"http://127.0.0.1:{port}/a2a/v1",
                    "protocolBinding": "HTTP+JSON",
                    "protocolVersion": "1.0",
                }
            ],
            "defaultInputModes": ["application/json"],
            "defaultOutputModes": ["application/json"],
            "skills": [{key: skill[key] for key in fields if key in skill} for skill in skills],
        }

    @app.post("/a2a/v1/message:send")
    def message_send(body: dict, request: Request):
        if protection and not secrets.compare_digest(
            request.headers.get(protection["header"], ""), protection["secret"]
        ):
            raise HTTPException(status_code=401, detail=f"missing or wrong {protection['header']}")

        skill = body.get("skill") or skills[0]["id"]
        data = body["message"]["parts"][0].get("data") or {}

        return {"message": {"role": "ROLE_AGENT", "parts": [{"data": funcs[skill](data)}]}}

    return app
