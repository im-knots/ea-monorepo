"use client";

import { useState } from "react";
import {
  Button,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
  Paper,
  Collapse,
  Box,
  Grid,
  Chip,
  IconButton,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";
import { KeyboardArrowDown, KeyboardArrowUp } from "@mui/icons-material";

interface Job {
  id: string;
  job_name: string;
  created_time: string;
  last_active?: string;
  status: string;
  agent_id: string;
  nodes?: Node[];
}

interface Node {
  alias: string;
  type?: string;
  output?: string;
  status?: string;
  lastUpdated?: string;
}

export default function JobListClient({
  initialJobs,
  agentId,
  userId,
}: {
  initialJobs: Job[];
  agentId: string;
  userId: string;
}) {
  const [jobs] = useState<Job[]>(initialJobs);
  const [expandedJobId, setExpandedJobId] = useState<string | null>(null);

  const handleExpand = (jobId: string) => {
    setExpandedJobId(expandedJobId === jobId ? null : jobId);
  };

  const getStatusColor = (
    status: string | undefined
  ): "success" | "warning" | "info" | "error" | "default" => {
    switch (status?.toLowerCase()) {
      case "completed":
      case "complete":
        return "success";
      case "pending":
        return "warning";
      case "executing":
        return "info";
      case "error":
        return "error";
      default:
        return "default";
    }
  };

  return (
    <Paper elevation={3} sx={{ p: 2, bgcolor: "grey.900", color: "white" }}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
        <Typography variant="h6">Jobs</Typography>
        <Button
          variant="outlined"
          size="small"
          startIcon={<RefreshIcon />}
          onClick={() => location.reload()}
        >
          Refresh
        </Button>
      </Box>

      {jobs.length > 0 ? (
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell />
              <TableCell>Job Name</TableCell>
              <TableCell>Created</TableCell>
              <TableCell>Last Active</TableCell>
              <TableCell>Status</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {jobs.map((job) => (
              <>
                <TableRow hover key={job.id}>
                  <TableCell>
                    <IconButton size="small" onClick={() => handleExpand(job.id)}>
                      {expandedJobId === job.id ? <KeyboardArrowUp /> : <KeyboardArrowDown />}
                    </IconButton>
                  </TableCell>
                  <TableCell
                    onClick={() => handleExpand(job.id)}
                    sx={{ cursor: "pointer", fontWeight: "medium" }}
                  >
                    {job.job_name}
                  </TableCell>
                  <TableCell>{new Date(job.created_time).toLocaleString()}</TableCell>
                  <TableCell>
                    {job.last_active ? new Date(job.last_active).toLocaleString() : "N/A"}
                  </TableCell>
                  <TableCell>
                    <Chip label={job.status} color={getStatusColor(job.status)} size="small" />
                  </TableCell>
                </TableRow>

                <TableRow key={`${job.id}-expanded`}>
                  <TableCell colSpan={5} sx={{ p: 0, border: 0 }}>
                    <Collapse in={expandedJobId === job.id} timeout="auto" unmountOnExit>
                      <Box sx={{ margin: 2 }}>
                        <Typography variant="subtitle1" gutterBottom>
                          Node Outputs
                        </Typography>
                        {job.nodes?.length ? (
                          <Grid container spacing={2}>
                            {job.nodes.map((node) => (
                              <Grid item xs={12} sm={6} md={4} key={node.alias}>
                                <Paper elevation={1} sx={{ p: 2, position: "relative", bgcolor: "grey.800" }}>
                                  <Chip
                                    label={node.status || "unknown"}
                                    color={getStatusColor(node.status)}
                                    size="small"
                                    sx={{ position: "absolute", top: 8, right: 8 }}
                                  />
                                  <Typography variant="body2" fontWeight="bold" color="grey.200">
                                    {node.alias}
                                  </Typography>
                                  <Typography variant="caption" color="grey.400">
                                    {node.lastUpdated
                                      ? `Updated: ${new Date(node.lastUpdated).toLocaleString()}`
                                      : "No timestamp"}
                                  </Typography>
                                  <Box
                                    component="pre"
                                    sx={{
                                      bgcolor: "grey.900",
                                      color: "grey.300",
                                      p: 1,
                                      borderRadius: 1,
                                      maxHeight: 120,
                                      overflow: "auto",
                                      fontSize: "0.75rem",
                                      mt: 1,
                                      whiteSpace: "pre-wrap",     // Ensures content wraps to the next line
                                      wordWrap: "break-word",     // Breaks long words/strings properly
                                    }}
                                  >
                                    {node.output ? JSON.stringify(JSON.parse(node.output), null, 2) : "No output"}
                                  </Box>
                                </Paper>
                              </Grid>
                            ))}
                          </Grid>
                        ) : (
                          <Typography variant="body2" color="grey.400">
                            No node outputs available.
                          </Typography>
                        )}
                      </Box>
                    </Collapse>
                  </TableCell>
                </TableRow>
              </>
            ))}
          </TableBody>
        </Table>
      ) : (
        <Typography variant="body2" color="grey.400">
          No jobs found for this agent.
        </Typography>
      )}
    </Paper>
  );
}
