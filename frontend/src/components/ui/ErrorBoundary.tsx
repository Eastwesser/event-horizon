import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Button } from './Button';

interface Props {
  children: ReactNode;
  /** Optional label for which page/section crashed. */
  label?: string;
}

interface State {
  error: Error | null;
}

/**
 * Catches render crashes (e.g. null.length) so the user sees a recovery UI
 * instead of a black screen.
 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('ErrorBoundary caught:', this.props.label, error, info);
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex min-h-[40vh] flex-col items-center justify-center gap-4 px-6 text-center">
          <p className="font-display text-xl font-semibold text-text-primary">Что-то пошло не так</p>
          <p className="max-w-md text-sm text-text-secondary">
            Не удалось показать эту страницу. Можно вернуться на главную и попробовать снова.
          </p>
          <Button
            variant="secondary"
            onClick={() => {
              this.setState({ error: null });
              window.location.assign('/');
            }}
          >
            На главную
          </Button>
        </div>
      );
    }
    return this.props.children;
  }
}
