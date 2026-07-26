import Link from "next/link";
import DecisionActions from "./DecisionActions";
import { decisionLabels, listAwaitingDecision } from "@/lib/api";

// Always render at request time — this page needs live data from apps/api,
// which is not reachable during `next build`'s static generation pass.
export const dynamic = "force-dynamic";

export default async function DecisionsPage() {
  const awaiting = await listAwaitingDecision();

  return (
    <div>
      <h1>Ждут решения</h1>

      {awaiting.length === 0 ? (
        // An empty list is a normal state of the system, not an error and not
        // an unexplained blank screen: nothing is waiting on a human.
        <p>Задач, ожидающих решения, нет.</p>
      ) : (
        <ul>
          {awaiting.map((entry) => (
            <li key={`${entry.task.projectId}/${entry.task.id}`}>
              <Link
                href={`/projects/${entry.task.projectId}/tasks/${entry.task.id}`}
              >
                {entry.task.id}: {entry.task.title}
              </Link>{" "}
              — <span>{decisionLabels[entry.decision]}</span>{" "}
              <span>({entry.task.projectId})</span>{" "}
              <DecisionActions awaiting={entry} />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
