import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import AgentRow from "./AgentRow";
import { fetchJobsForAgent } from "./JobActions";
import { Paper, Box, Typography, Button, Table, TableHead, TableRow, TableCell, TableBody } from "@mui/material";

const AGENT_MANAGER_URL = "http://api.erulabs.local/agent-manager/api/v1/agents";

export default async function AgentTable() {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;

  if (!token) {
    return redirect("/login");
  }

  const payload = JSON.parse(atob(token.split(".")[1]));
  const userId = payload?.sub;

  const agentsRes = await fetch(AGENT_MANAGER_URL, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });

  if (!agentsRes.ok) {
    throw new Error("Failed to fetch agents");
  }

  const agents = await agentsRes.json();

  const detailedAgents = await Promise.all(
    agents.map(async (agent: any) => {
      const [detailRes, jobs] = await Promise.all([
        fetch(`${AGENT_MANAGER_URL}/${agent.id}`, {
          headers: { Authorization: `Bearer ${token}` },
          cache: "no-store",
        }),
        fetchJobsForAgent(agent.id, userId),
      ]);

      if (!detailRes.ok) {
        throw new Error("Failed to fetch agent details");
      }

      const details = await detailRes.json();

      return {
        ...agent,
        ...details,
        jobs,
        jobCount: jobs.length,
      };
    })
  );

  return (
    <Box sx={{ padding: 4 }}>
      <Paper elevation={4} sx={{ padding: 3 }}>
        <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
          <Typography variant="h5" component="h2">
            My Agents
          </Typography>
          <Box display="flex" gap={1}>
            <Button variant="contained" color="primary" href="/agent-builder">
              + Create Agent
            </Button>
            <Button variant="contained" color="primary" href="/node-builder">
              + Create Node
            </Button>
          </Box>
        </Box>

        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Agent Name</TableCell>
              <TableCell align="center">Nodes</TableCell>
              <TableCell align="center">Jobs</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {detailedAgents.map((agent) => (
              <AgentRow
                key={agent.id}
                agent={agent}
                userId={userId}
                initialJobs={agent.jobs}
              />
            ))}
          </TableBody>
        </Table>
      </Paper>
    </Box>
  );
}
