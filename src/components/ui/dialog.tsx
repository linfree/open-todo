import * as React from "react"
import { cn } from "../../lib/utils"
import { X } from "lucide-react"

// --- Context ---

interface DialogContextValue {
  isAnimating: boolean
  closeOnBackdrop: boolean
  onOpenChange: (open: boolean) => void
}

const DialogContext = React.createContext<DialogContextValue | null>(null)

function useDialogContext() {
  const ctx = React.useContext(DialogContext)
  if (!ctx) {
    throw new Error("Dialog components must be used within <Dialog>")
  }
  return ctx
}

// --- Dialog ---

interface DialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  /** Whether clicking the backdrop closes the dialog. Defaults to true. */
  closeOnBackdrop?: boolean
  children: React.ReactNode
}

const Dialog = ({
  open,
  onOpenChange,
  closeOnBackdrop = true,
  children,
}: DialogProps) => {
  const dialogRef = React.useRef<HTMLDivElement>(null)
  const previousActiveElement = React.useRef<Element | null>(null)
  const closeTimerRef = React.useRef<ReturnType<typeof setTimeout> | undefined>(undefined)
  const [isMounted, setIsMounted] = React.useState(false)
  const [isAnimating, setIsAnimating] = React.useState(false)

  // Exit animation: keep DOM mounted briefly for close animation
  React.useEffect(() => {
    if (open) {
      clearTimeout(closeTimerRef.current)
      setIsMounted(true)
      setIsAnimating(true)
      previousActiveElement.current = document.activeElement
    } else {
      setIsAnimating(false)
      closeTimerRef.current = setTimeout(() => {
        setIsMounted(false)
        if (previousActiveElement.current instanceof HTMLElement) {
          previousActiveElement.current.focus()
        }
      }, 200)
    }
    return () => {
      clearTimeout(closeTimerRef.current)
    }
  }, [open])

  // Keyboard: Escape to close
  React.useEffect(() => {
    if (!open) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault()
        onOpenChange(false)
      }
    }
    document.addEventListener("keydown", handleKeyDown)
    return () => document.removeEventListener("keydown", handleKeyDown)
  }, [open, onOpenChange])

  // Focus trap: focus first focusable element on open, prevent Tab from escaping
  React.useEffect(() => {
    if (!open) return
    const dialog = dialogRef.current
    if (!dialog) return

    const getFocusable = () =>
      dialog.querySelectorAll<HTMLElement>(
        'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
      )

    // Focus the first focusable element on open
    const focusable = getFocusable()
    if (focusable.length > 0) {
      focusable[0].focus()
    }

    const handleTabKey = (e: KeyboardEvent) => {
      if (e.key !== "Tab") return
      const els = getFocusable()
      if (els.length === 0) return
      const first = els[0]
      const last = els[els.length - 1]
      if (e.shiftKey && document.activeElement === first) {
        last.focus()
        e.preventDefault()
      } else if (!e.shiftKey && document.activeElement === last) {
        first.focus()
        e.preventDefault()
      }
    }

    document.addEventListener("keydown", handleTabKey)
    return () => document.removeEventListener("keydown", handleTabKey)
  }, [open])

  const ctxValue = React.useMemo(
    () => ({ isAnimating, closeOnBackdrop, onOpenChange }),
    [isAnimating, closeOnBackdrop, onOpenChange]
  )

  if (!isMounted) return null

  return (
    <DialogContext.Provider value={ctxValue}>
      <div className="fixed inset-0 z-50 flex items-center justify-center">
        <div
          className={cn(
            "fixed inset-0 bg-black/40 backdrop-blur-sm duration-200",
            isAnimating ? "animate-in fade-in-0" : "animate-out fade-out-0"
          )}
          onClick={() => closeOnBackdrop && onOpenChange(false)}
        />
        <div ref={dialogRef} className="relative z-50">
          {children}
        </div>
      </div>
    </DialogContext.Provider>
  )
}

// --- DialogContent ---

interface DialogContentProps extends React.HTMLAttributes<HTMLDivElement> {
  "aria-labelledby"?: string
  "aria-describedby"?: string
}

const DialogContent = React.forwardRef<HTMLDivElement, DialogContentProps>(
  (
    {
      className,
      children,
      "aria-labelledby": ariaLabelledby,
      "aria-describedby": ariaDescribedby,
      ...props
    },
    ref
  ) => {
    const { isAnimating } = useDialogContext()

    return (
      <div
        ref={ref}
        role="dialog"
        aria-modal="true"
        aria-labelledby={ariaLabelledby}
        aria-describedby={ariaDescribedby}
        className={cn(
          "fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border border-border/50 bg-card shadow-xl duration-200",
          isAnimating
            ? "animate-in fade-in-0 zoom-in-95 slide-in-from-left-1/2 slide-in-from-top-[48%]"
            : "animate-out fade-out-0 zoom-out-95 slide-out-to-left-1/2 slide-out-to-top-[48%]",
          // 移动端全屏显示
          "sm:rounded-xl rounded-none sm:p-6 p-4 h-full sm:h-auto sm:max-h-[90vh] max-h-screen overflow-y-auto scrollbar-hide",
          className
        )}
        {...props}
      >
        {children}
      </div>
    )
  }
)
DialogContent.displayName = "DialogContent"

// --- DialogHeader ---

const DialogHeader = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={cn(
      "flex flex-col space-y-1.5 text-center sm:text-left",
      className
    )}
    {...props}
  />
)
DialogHeader.displayName = "DialogHeader"

// --- DialogTitle ---

const DialogTitle = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => (
  <h2
    ref={ref}
    className={cn(
      "text-lg font-semibold leading-none tracking-tight text-foreground",
      className
    )}
    {...props}
  />
))
DialogTitle.displayName = "DialogTitle"

// --- DialogDescription ---

const DialogDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <p
    ref={ref}
    className={cn("text-sm text-muted-foreground", className)}
    {...props}
  />
))
DialogDescription.displayName = "DialogDescription"

// --- DialogFooter ---

const DialogFooter = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={cn(
      "flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2",
      className
    )}
    {...props}
  />
)
DialogFooter.displayName = "DialogFooter"

// --- DialogClose ---

const DialogClose = React.forwardRef<
  HTMLButtonElement,
  React.ButtonHTMLAttributes<HTMLButtonElement>
>(({ onClick, className, ...props }, ref) => {
  const { onOpenChange } = useDialogContext()

  return (
    <button
      ref={ref}
      onClick={onClick ?? (() => onOpenChange(false))}
      className={cn(
        "absolute right-4 top-4 rounded-lg opacity-70 ring-offset-background transition-all hover:opacity-100 hover:bg-accent focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none p-2 cursor-pointer text-foreground hover:text-accent-foreground",
        className
      )}
      {...props}
    >
      <X className="h-4 w-4" />
      <span className="sr-only">Close</span>
    </button>
  )
})
DialogClose.displayName = "DialogClose"

export {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
}
