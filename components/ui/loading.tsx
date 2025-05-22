import { Loader2 } from "lucide-react"

interface LoadingProps {
  message?: string
  className?: string
}

export function Loading({ message = "Loading...", className }: LoadingProps) {
  return (
    <div className={`flex flex-col items-center justify-center p-8 ${className}`}>
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      <p className="mt-2 text-sm text-muted-foreground">{message}</p>
    </div>
  )
}
