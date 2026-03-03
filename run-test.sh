echo "Starting agent Tests"
rm -rf kanshi-binary
go build -o kanshi-agent cmd/agent/main.go

read -p "Enter number of agents to run: " num_agents

pids=""

cleanup() {
  echo "Stopping agents..."
  kill $pids 2>/dev/null
  echo "All agents stopped"
  rm -rf kanshi-binary
}

trap cleanup EXIT

for i in $(seq 1 "$num_agents"); do
  mkdir -p "kanshi-binary/agent$i"
  cp kanshi-agent "kanshi-binary/agent$i/"
  (cd "./kanshi-binary/agent$i" && ./kanshi-agent) &
  pids="$pids $!"
done

rm kanshi-agent

echo "Agent PIDs:$pids"
echo "$pids" > kanshi-binary/agent.pids

# shellcheck disable=SC2154
read -p "Type 'stop' to stop agents: " $input

cleanup