/** Colores del diseño de referencia — navy, teal, pink, violet (solo Pablo). */
export type PersonPalette = {
  solid: string;
  gradient: string;
  legend: string;
};

const NAMED: Record<string, PersonPalette> = {
  diego: {
    solid: "#059669",
    gradient: "linear-gradient(135deg, #10b981 0%, #059669 100%)",
    legend: "#10b981",
  },
  enrique: {
    solid: "#ec4899",
    gradient: "linear-gradient(135deg, #f472b6 0%, #ec4899 100%)",
    legend: "#ec4899",
  },
  gerardo: {
    solid: "#14b8a6",
    gradient: "linear-gradient(135deg, #2dd4bf 0%, #14b8a6 100%)",
    legend: "#14b8a6",
  },
  pablo: {
    solid: "#8b5cf6",
    gradient: "linear-gradient(135deg, #a78bfa 0%, #8b5cf6 100%)",
    legend: "#8b5cf6",
  },
};

const FALLBACK: PersonPalette[] = [
  {
    solid: "#059669",
    gradient: "linear-gradient(135deg, #10b981 0%, #059669 100%)",
    legend: "#10b981",
  },
  {
    solid: "#14b8a6",
    gradient: "linear-gradient(135deg, #2dd4bf 0%, #14b8a6 100%)",
    legend: "#14b8a6",
  },
  {
    solid: "#3730a3",
    gradient: "linear-gradient(135deg, #4f46e5 0%, #3730a3 100%)",
    legend: "#4f46e5",
  },
  {
    solid: "#ec4899",
    gradient: "linear-gradient(135deg, #f472b6 0%, #ec4899 100%)",
    legend: "#ec4899",
  },
  {
    solid: "#8b5cf6",
    gradient: "linear-gradient(135deg, #a78bfa 0%, #8b5cf6 100%)",
    legend: "#8b5cf6",
  },
];

function hash(person: string): number {
  let h = 0;
  const k = person.trim().toLowerCase();
  for (let i = 0; i < k.length; i++) h = k.charCodeAt(i) + ((h << 5) - h);
  return Math.abs(h);
}

export function getPersonPalette(person: string): PersonPalette {
  const key = person.trim().toLowerCase();
  return NAMED[key] ?? FALLBACK[hash(person) % FALLBACK.length]!;
}

export function personColor(person: string): string {
  return getPersonPalette(person).solid;
}

export function personLegendColor(person: string): string {
  return getPersonPalette(person).legend;
}

export function personEventGradient(person: string): string {
  return getPersonPalette(person).gradient;
}
