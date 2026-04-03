import { Moon, Sun } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useTheme } from '@/context/theme-provider'

export function ThemeToggle() {
  const { resolvedTheme, setTheme, theme } = useTheme()

  const cycle = () => {
    const next = theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light'
    setTheme(next)
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={cycle}
      className="relative size-8"
    >
      <Sun className="size-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
      <Moon className="absolute size-4 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
      <span className="sr-only">
        {resolvedTheme === 'dark' ? 'Mudar para light' : 'Mudar para dark'}
      </span>
    </Button>
  )
}
