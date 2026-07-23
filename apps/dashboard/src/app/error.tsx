"use client";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div>
      <p>Не удалось загрузить данные: {error.message}</p>
      <button onClick={() => reset()}>Попробовать снова</button>
    </div>
  );
}
