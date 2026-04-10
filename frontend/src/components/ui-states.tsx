interface StateProps {
  message: string;
}

export function LoadingState({ message }: StateProps) {
  return <p className="panel text-sm text-slate-600">{message}</p>;
}

export function EmptyState({ message }: StateProps) {
  return (
    <div className="panel">
      <p className="text-slate-600">{message}</p>
    </div>
  );
}

export function ErrorState({ message }: StateProps) {
  return <p className="border border-red-200 bg-red-50 p-4 text-sm text-red-700">Ошибка: {message}</p>;
}
