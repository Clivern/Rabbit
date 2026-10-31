#!/usr/bin/env bash
# Run all demo agents at once. Ctrl+C stops them and deregisters their instances.

cd "$(dirname "$0")"

export SLOTH_URL=http://localhost:8000
export SLOTH_WORKSPACE="<workspace-id>"
export SLOTH_ACCESS_KEY="<access-key>"
export SLOTH_AGENT_SECRET="<assistant-secret>"  # any random value, min 8 characters; the assistant issues it as its identity on boot
export MODEL_API_KEY="<model-key>"
export ASSISTANT_MODEL=muse-spark-1.3-contributor
export PYTHONUNBUFFERED=1

# Keep real values in run.local.sh, which git ignores.
[ -f run.local.sh ] && source run.local.sh

# Four instances per agent, assistant stays single. The third field is the
# pretty instance id Sloth shows (dc in the middle); the address stays
# 127.0.0.1 so Sloth can reach these local demo processes.
agents=(
  assistant:9200:
  weather:9201:weather01.ams01.xai.com
  weather:9211:weather02.ams01.xai.com
  weather:9221:weather03.fra01.xai.com
  weather:9231:weather04.fra01.xai.com
  stock:9202:stock01.ams01.xai.com
  stock:9212:stock02.ams01.xai.com
  stock:9222:stock03.fra01.xai.com
  stock:9232:stock04.fra01.xai.com
  users:9203:users01.ams01.xai.com
  users:9213:users02.ams01.xai.com
  users:9223:users03.fra01.xai.com
  users:9233:users04.fra01.xai.com
)

for agent in "${agents[@]}"; do
  IFS=':' read -r name port instance <<< "$agent"
  tag=${instance:-$name:$port}
  INSTANCE="$instance" PORT="$port" uv run uvicorn "$name:app" --port "$port" > >(sed "s/^/[$tag] /") 2>&1 &
done

trap 'kill $(jobs -p) 2>/dev/null' INT TERM
wait
