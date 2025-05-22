"use client"

import { AlertCircle } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"

interface ErrorDisplayProps {
  title?: string
  message: string
  retry?: () => void
}

export function ErrorDisplay({ title = "Error", message, retry }: ErrorDisplayProps) {
  return (
    <Alert variant="destructive" className="my-4">
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription className="flex flex-col gap-2">
        <p>{message}</p>
        {retry && (
          <Button variant="outline" size="sm" onClick={retry} className="w-fit">
            Retry
          </Button>
        )}
      </AlertDescription>
    </Alert>
  )
}
