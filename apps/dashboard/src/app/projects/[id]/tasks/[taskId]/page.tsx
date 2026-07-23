import { getTask } from "@/lib/api";

// Always render at request time — see TASK-075's Отчёт.
export const dynamic = "force-dynamic";

export default async function TaskPage({
  params,
}: {
  params: Promise<{ id: string; taskId: string }>;
}) {
  const { id, taskId } = await params;
  const task = await getTask(id, taskId);

  return (
    <div>
      <h1>
        {task.id} — {task.title}
      </h1>
      <dl>
        <dt>Проект</dt>
        <dd>{task.projectId}</dd>
        <dt>Тип</dt>
        <dd>{task.type}</dd>
        <dt>Состояние</dt>
        <dd>{task.state}</dd>
        <dt>Scope</dt>
        <dd>{task.scope || "—"}</dd>
        <dt>Критерии приёмки</dt>
        <dd>
          {task.acceptanceCriteria.length === 0 ? (
            "—"
          ) : (
            <ul>
              {task.acceptanceCriteria.map((criterion) => (
                <li key={criterion}>{criterion}</li>
              ))}
            </ul>
          )}
        </dd>
        <dt>Обновлено</dt>
        <dd>{new Date(task.updatedAt).toLocaleString("ru-RU")}</dd>
      </dl>
    </div>
  );
}
