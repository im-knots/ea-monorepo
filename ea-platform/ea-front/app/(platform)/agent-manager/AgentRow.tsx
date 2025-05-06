"use client";

import { useState, useTransition, useEffect } from "react";
import JobListClient from "./JobListClient";
import { startAgentJob, deleteAgent } from "./AgentActions";
import { Button, TableCell, TableRow, Collapse, Box } from "@mui/material";
import { KeyboardArrowDown, KeyboardArrowUp } from "@mui/icons-material";

export default function AgentRow({ agent, userId, initialJobs }: { agent: any; userId: string; initialJobs: any[] }) {
  const [expanded, setExpanded] = useState(false);
  const [isStarting, startTransition] = useTransition();
  const [isDeleting, deleteTransition] = useTransition();

  useEffect(() => {
    const storedExpanded = sessionStorage.getItem(`agent-expanded-${agent.id}`);
    setExpanded(storedExpanded === "true");
  }, [agent.id]);

  const handleExpandToggle = () => {
    setExpanded((prev) => {
      const newVal = !prev;
      sessionStorage.setItem(`agent-expanded-${agent.id}`, newVal.toString());
      return newVal;
    });
  };

  const handleStart = () => {
    startTransition(async () => {
      await startAgentJob(agent.id, userId);
      setExpanded(true);
      sessionStorage.setItem(`agent-expanded-${agent.id}`, "true");
    });
  };

  const handleDelete = () => {
    deleteTransition(async () => {
      await deleteAgent(agent.id);
      location.reload();
    });
  };

  return (
    <>
      <TableRow hover sx={{ cursor: "pointer" }} onClick={handleExpandToggle}>
        <TableCell>{agent.name}</TableCell>
        <TableCell align="center">{agent.nodes.length}</TableCell>
        <TableCell align="center">{agent.jobCount}</TableCell>
        <TableCell align="center">
          <Box display="flex" justifyContent="center" gap={1}>
            <Button
              variant="contained"
              color="success"
              size="small"
              disabled={isStarting}
              onClick={(e) => { e.stopPropagation(); handleStart(); }}
            >
              {isStarting ? "Starting..." : "Start"}
            </Button>
            <Button variant="contained" color="primary" size="small">Modify</Button>
            <Button
              variant="contained"
              color="error"
              size="small"
              disabled={isDeleting}
              onClick={(e) => { e.stopPropagation(); handleDelete(); }}
            >
              {isDeleting ? "Deleting..." : "Delete"}
            </Button>
            <Button size="small">
              {expanded ? <KeyboardArrowUp /> : <KeyboardArrowDown />}
            </Button>
          </Box>
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell colSpan={4} sx={{ padding: 0, borderBottom: "none" }}>
          <Collapse in={expanded} timeout="auto" unmountOnExit>
            <Box sx={{ margin: 1 }}>
              <JobListClient initialJobs={initialJobs} agentId={agent.id} userId={userId} />
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  );
}
