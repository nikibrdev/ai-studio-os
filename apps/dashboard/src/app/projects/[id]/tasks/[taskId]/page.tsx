import { getArtifact, getTask } from "@/lib/api";

// Always render at request time — see TASK-075's Отчёт.
export const dynamic = "force-dynamic";

export default async function TaskPage({
  params,
}: {
  params: Promise<{ id: string; taskId: string }>;
}) {
  const { id, taskId } = await params;
  const task = await getTask(id, taskId);

  // The QA report the human reads before the acceptance decision (TASK-100).
  // Fetched separately because the task carries only its identifier — the report
  // lives in the artifact.
  const qaReport =
    task.qaReportId === "" ? null : await getArtifact(task.qaReportId);

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

      <section>
        <h2>Отчёт QA</h2>
        {qaReport === null ? (
          // Explicitly "not checked", never a blank space: an empty area next to
          // an acceptance decision reads as "nothing to report", and a human
          // would merge on the strength of a check that never happened.
          <p>Проверка ещё не выполнялась — отчёта нет.</p>
        ) : (
          <>
            <p>
              Автор: {qaReport.author},{" "}
              {new Date(qaReport.createdAt).toLocaleString("ru-RU")}
            </p>
            <pre>{qaReport.payload}</pre>
          </>
        )}
      </section>
    </div>
  );
}
