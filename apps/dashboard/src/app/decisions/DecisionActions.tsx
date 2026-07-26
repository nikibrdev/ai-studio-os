"use client";

import { useState, useTransition } from "react";
import {
  acceptDefinitionOfReady,
  acceptTask,
  returnForRework,
  type ActionResult,
} from "@/lib/actions";
import type { AwaitingDecision } from "@/lib/api";

// DecisionActions — первый мутирующий UI в проекте: до этого apps/dashboard был
// целиком read-only. Отсюда и объём внимания к отказам и обновлению данных: это
// новая для проекта категория кода, а не просто новые кнопки.
export default function DecisionActions({
  awaiting,
}: {
  awaiting: AwaitingDecision;
}) {
  const { decision, task } = awaiting;
  const [pending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);

  function run(action: () => Promise<ActionResult>) {
    setError(null);
    startTransition(async () => {
      const result = await action();
      if (!result.ok) {
        // Показываем рядом с кнопкой: страница остаётся на месте, повторить
        // можно сразу. Состояние задачи при этом не изменилось — выдавать
        // прежнее за новое нельзя, поэтому revalidate делается только при успехе.
        setError(result.error);
      }
    });
  }

  if (decision === "definition-of-ready") {
    return (
      <span>
        <button
          type="button"
          disabled={pending}
          onClick={() =>
            run(() => acceptDefinitionOfReady(task.projectId, task.id))
          }
        >
          {/* Коротко: рядом уже стоит название решения из decisionLabels,
              повторять его на кнопке — лишний шум. */}
          {pending ? "Применяется…" : "Принять"}
        </button>
        {error !== null && <span role="alert"> Не удалось: {error}</span>}
      </span>
    );
  }

  return (
    <span>
      {/* Предупреждение до нажатия, а не после: приёмочное решение сливает Pull
          Request и в рамках задачи необратимо (workflow.md, ADR-008). */}
      <span>
        Принятие сольёт{" "}
        {task.pullRequestId !== ""
          ? `Pull Request ${task.pullRequestId} в ${task.repository}`
          : "Pull Request"}{" "}
        в основную ветку — в рамках задачи это необратимо.
      </span>{" "}
      <button
        type="button"
        disabled={pending}
        onClick={() => run(() => acceptTask(task.projectId, task.id))}
      >
        {pending ? "Применяется…" : "Принять и слить"}
      </button>{" "}
      <button
        type="button"
        disabled={pending}
        onClick={() => run(() => returnForRework(task.projectId, task.id))}
      >
        Вернуть на доработку
      </button>
      {error !== null && <span role="alert"> Не удалось: {error}</span>}
    </span>
  );
}
