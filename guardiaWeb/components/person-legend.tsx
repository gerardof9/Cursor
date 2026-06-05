import { personLegendColor } from "@/lib/person-colors";
import { cn } from "@/lib/utils";

export function PersonLegend({
  people,
  className,
}: {
  people: string[];
  className?: string;
}) {
  if (people.length === 0) return null;

  return (
    <div className={cn("flex flex-wrap gap-2", className)}>
      {people.map((person) => (
        <span key={person} className="legend-chip">
          <span
            className="legend-dot"
            style={{ backgroundColor: personLegendColor(person) }}
          />
          {person}
        </span>
      ))}
    </div>
  );
}
