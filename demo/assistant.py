"""Assistant agent with a chat page. uv run uvicorn assistant:app --port 9200

It answers a question by calling other agents through the Sloth gateway:

1. list the skills it may call: GET /agents/catalog/skills
2. ask an LLM which skills to call and with what data
3. call each one through Sloth: POST /agents/skills/{skill}/message:send
4. ask the LLM to turn the results into one short answer

POST /ask streams each step and the answer text to the chat page as
server-sent events; the A2A message:send endpoint still gets one answer.

Open http://127.0.0.1:9200 for the chat page. Besides the registration variables
in util.py it needs SLOTH_AGENT_SECRET, any random value it issues as its own
api_key identity on boot, and MODEL_API_KEY. ASSISTANT_MODEL picks the model
and MODEL_API_URL picks the endpoint: the Meta responses API when the URL ends
with /responses (default), OpenAI chat completions otherwise (e.g. OpenRouter).
"""

import json
import os
from pathlib import Path

import httpx
from fastapi import FastAPI
from fastapi.responses import HTMLResponse, StreamingResponse

from util import INSTANCE, URL, WORKSPACE, lifespan

NAME = "assistant"
VERSION = "1.0.0"
PORT = int(os.environ.get("PORT", "9200"))
SECRET = os.environ.get("SLOTH_AGENT_SECRET", "")
HEADERS = {"X-Source-Agent": NAME, "X-Source-Instance": INSTANCE or f"{NAME}-{PORT}", "X-API-Key": SECRET}
MODEL_API_KEY = os.environ.get("MODEL_API_KEY", "")
MODEL = os.environ.get("ASSISTANT_MODEL", "muse-spark-1.3-contributor")
MODEL_API_URL = os.environ.get("MODEL_API_URL", "https://api.meta.ai/v1/responses")


def is_meta_api():
    """Whether the model endpoint speaks the Meta responses API; anything else uses OpenAI chat completions."""
    return MODEL_API_URL.rstrip("/").endswith("/responses")
GATEWAY = f"{URL}/api/v1/workspaces/{WORKSPACE}/agents"
MAX_CALLS = 3
PAGE = (Path(__file__).parent / "index.html").read_text()

PLAN = """You route a user's question to agent skills.
You get the question and the skills you may call, each with an id, a name and a
description that ends with an example of its data. Reply with JSON only:
{"calls": [{"skill": "<skill id>", "data": {...}}]}
Use only the listed skill ids, at most 3 calls, and fill data from the question
using the same keys as the example. Reply {"calls": []} when no skill helps."""

ANSWER = """You answer the user's question in two or three friendly sentences.
Base live data such as weather, stock prices and user details only on the agent
results you are given, and never invent it. Never name skills, tools, calls or
HTTP statuses; just answer from the results. When a result is marked
unavailable, say plainly in one sentence which information you could not get.
When there are no results because no skill fits, answer general knowledge
questions yourself and say which kinds of questions your agents can help with."""

app = FastAPI(lifespan=lifespan(NAME, VERSION, PORT, calls=["weather", "stock", "users"], secret=SECRET))


def prompt(system, content):
    if is_meta_api():
        return {
            "model": MODEL,
            "instructions": system,
            "input": [
                {
                    "role": "user",
                    "content": [{"type": "input_text", "text": json.dumps(content)}],
                }
            ],
            "reasoning": {"effort": "low"},
            "stream": True,
        }

    return {
        "model": MODEL,
        "messages": [
            {"role": "system", "content": system},
            {"role": "user", "content": json.dumps(content)},
        ],
        "stream": True,
    }


async def chat_stream(client, system, content):
    """Yield answer text pieces from the model API as they arrive."""
    headers = {"Authorization": f"Bearer {MODEL_API_KEY}"}

    if not is_meta_api():
        headers["HTTP-Referer"] = "https://github.com/clivern/sloth"
        headers["X-Title"] = "Sloth demo"

    async with client.stream(
        "POST",
        MODEL_API_URL,
        headers=headers,
        json=prompt(system, content),
    ) as response:
        response.raise_for_status()

        async for line in response.aiter_lines():
            if not line.startswith("data: ") or line == "data: [DONE]":
                continue

            event = json.loads(line[6:])

            if is_meta_api():
                if event.get("type") == "response.output_text.delta":
                    yield event.get("delta", "")
            else:
                for choice in event.get("choices", []):
                    yield choice.get("delta", {}).get("content", "")


async def chat(client, system, content):
    """Full response text from the Meta API, collected from its stream."""
    return "".join([part async for part in chat_stream(client, system, content)])


async def fetch_skills(client):
    """Skills Sloth lets the assistant call, with the listing status."""
    response = await client.get(f"{GATEWAY}/catalog/skills", headers=HEADERS)
    skills = response.json().get("skills", []) if response.is_success else []

    return response.status_code, skills


def answer_fact(item):
    """Inner payload for the answer model; failures become a plain unavailable note."""
    if item["status"] >= 400:
        return "unavailable: this information could not be retrieved"

    part = item["result"]["message"]["parts"][0]

    return part.get("data") or part.get("text")


async def run(question):
    """Yield ("step", step) and ("delta", text) events while answering a question through Sloth."""
    async with httpx.AsyncClient(timeout=60) as client:
        # 1. Sloth only lists skills an intention lets the assistant call.
        status, skills = await fetch_skills(client)
        yield ("step", {"step": "catalog", "status": status, "skills": [skill["id"] for skill in skills]})

        if status == 401:
            yield ("delta", "Sloth doesn't recognize my secret (401). Set SLOTH_AGENT_SECRET to any random value (min 8 characters) and restart me; if it still fails, revoke the assistant's identities under Agents in Sloth and restart again.")
            return

        if not skills:
            yield ("delta", "I can't call any agent yet. Allow the assistant to call an agent under Intentions in Sloth.")
            return

        # 2. The LLM picks skills and their data from the question.
        choices = [{"id": s["id"], "name": s["name"], "description": s["description"]} for s in skills]
        plan = json.loads(await chat(client, PLAN, {"question": question, "skills": choices}))
        calls = [call for call in plan.get("calls", []) if call.get("skill") in {s["id"] for s in skills}][:MAX_CALLS]
        yield ("step", {"step": "plan", "calls": calls})

        # 3. Each call goes through Sloth, which checks intentions and picks a live instance.
        facts = []
        for call in calls:
            response = await client.post(
                f"{GATEWAY}/skills/{call['skill']}/message:send",
                headers=HEADERS,
                json={
                    "skill": call["skill"],
                    "message": {"role": "ROLE_USER", "parts": [{"data": call.get("data", {})}]},
                },
            )
            result = response.json()
            facts.append(answer_fact({"status": response.status_code, "result": result}))
            yield ("step", {"step": "call", "skill": call["skill"], "data": call.get("data", {}), "status": response.status_code, "result": result})

        # 4. The LLM turns the facts into one answer, streamed as it arrives.
        async for piece in chat_stream(client, ANSWER, {"question": question, "results": facts}):
            yield ("delta", piece)

        yield ("step", {"step": "answer", "model": MODEL})


async def ask(question):
    """Answer a question through Sloth and return the answer with every step taken."""
    steps = []
    answer = []

    async for kind, payload in run(question):
        if kind == "step":
            steps.append(payload)
        else:
            answer.append(payload)

    return {"answer": "".join(answer), "steps": steps}


@app.get("/", response_class=HTMLResponse)
def page():
    return PAGE


@app.get("/info")
async def info():
    """The skills Sloth lets the assistant call, for the page's setup card."""
    async with httpx.AsyncClient(timeout=10) as client:
        status, skills = await fetch_skills(client)

    return {"status": status, "skills": [skill["id"] for skill in skills]}


@app.post("/ask")
async def ask_action(body: dict):
    async def events():
        async for kind, payload in run(body["question"]):
            yield f"event: {kind}\ndata: {json.dumps(payload)}\n\n"

    return StreamingResponse(events(), media_type="text/event-stream")


@app.get("/healthz")
def healthz():
    return {"status": "ok"}


@app.get("/.well-known/agent-card.json")
def agent_card():
    return {
        "name": "Assistant",
        "description": "Answers questions by asking other agents through Sloth",
        "version": VERSION,
        "supportedInterfaces": [
            {
                "url": f"http://127.0.0.1:{PORT}/a2a/v1",
                "protocolBinding": "HTTP+JSON",
                "protocolVersion": "1.0",
            }
        ],
        "defaultInputModes": ["text/plain"],
        "defaultOutputModes": ["text/plain"],
        "skills": [
            {
                "id": "assistant.ask",
                "name": "Ask the assistant",
                "description": "Answer a question in plain language using the agents the assistant may call",
                "tags": ["assistant"],
                "examples": ["What's the weather in Cairo and how is NVDA doing?"],
            }
        ],
    }


@app.post("/a2a/v1/message:send")
async def message_send(body: dict):
    question = body["message"]["parts"][0].get("text", "")
    result = await ask(question)

    return {"message": {"role": "ROLE_AGENT", "parts": [{"text": result["answer"]}]}}
