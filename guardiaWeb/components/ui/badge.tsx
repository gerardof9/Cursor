import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const variants = cva(
  "inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-bold text-white",
  {
    variants: {
      variant: {
        success: "bg-emerald-600 shadow-sm",
      },
    },
    defaultVariants: { variant: "success" },
  },
);

export function Badge({
  className,
  variant,
  ...props
}: React.HTMLAttributes<HTMLSpanElement> & VariantProps<typeof variants>) {
  return <span className={cn(variants({ variant }), className)} {...props} />;
}
