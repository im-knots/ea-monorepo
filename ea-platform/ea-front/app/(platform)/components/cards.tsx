import { Box, Typography } from "@mui/material";

export default function CardWrapper() {
  const cards = [
    { title: "My compute credits", value: 0, type: "credits", units: "credits" },
    { title: "My Compute rate", value: 0, type: "compute_rate", units: "TFLOPS" },
    { title: "Inference jobs", value: 0, type: "inference_jobs", units: "inference jobs" },
    { title: "Training jobs", value: 0, type: "training_jobs", units: "training jobs" },
    { title: "Agents", value: 0, type: "agents", units: "agents" },
  ]
  return (
    <Box className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-5">
      {cards.map((card) => (
        <Card key={card.type} title={card.title} value={card.value} type={card.type} unit={card.units} />
      ))}
    </Box>
  )
}

export function Card({
  title,
  value,
  type,
  unit,
}: {
  title: string;
  value: number | string;
  type: string;
  unit: string;
}) {

  return (
    <div className="rounded-xl bg-slate-700 p-2 shadow-sm">
      <div className="flex p-6">
        <Typography variant="h5" className="ml-2 text-sm font-medium">{title}</Typography>
      </div>
      <p
        className={`truncate rounded-xl px-4 py-8 text-center text-2xl`}
      >
        {value} {unit}
      </p>
    </div>
  );
}