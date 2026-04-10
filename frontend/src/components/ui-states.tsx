interface StateProps {
  message: string;
}

export function LoadingState({ message }: StateProps) {
  return <p className="panel text-sm text-slate-700">{message}</p>;
}

export function EmptyState({ message }: StateProps) {
  return (
    <div className="panel">
      <p className="text-slate-700">{message}</p>
    </div>
  );
}

export function ErrorState({ message }: StateProps) {
  return <p className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">Ошибка: {message}</p>;
}

