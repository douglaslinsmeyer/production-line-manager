import { useState } from 'react';
import { ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/24/outline';
import { formatDate, formatRelativeTime, formatDuration } from '@/utils/formatters';
import StatusBadge from '@/components/common/StatusBadge';
import type { StatusChange } from '@/api/types';

interface StatusTimelineProps {
  history: StatusChange[];
}

export default function StatusTimeline({ history }: StatusTimelineProps) {
  const [expandedItems, setExpandedItems] = useState<Set<number>>(new Set());

  const toggleExpand = (index: number) => {
    const newExpanded = new Set(expandedItems);
    if (newExpanded.has(index)) {
      newExpanded.delete(index);
    } else {
      newExpanded.add(index);
    }
    setExpandedItems(newExpanded);
  };

  if (!history || history.length === 0) {
    return (
      <div className="flex items-center justify-center py-8">
        <p className="text-gray-500">No status changes recorded</p>
      </div>
    );
  }

  return (
    <div className="flow-root">
      <ul className="-mb-8">
        {history.map((change, index) => {
          // Calculate duration to next event (if not the last item)
          const duration = index < history.length - 1
            ? Math.abs(new Date(history[index + 1].time).getTime() - new Date(change.time).getTime())
            : null;

          return (
          <li key={`${change.time}-${index}`}>
            <div className="relative pb-8 flex">
              {/* Duration column (left side) */}
              <div className="w-20 flex-shrink-0 text-right pr-4">
                {duration && (
                  <span className="text-xs text-gray-500 font-medium">
                    {formatDuration(duration)}
                  </span>
                )}
              </div>

              {/* Timeline column (middle and right) */}
              <div className="relative flex-1">
                {/* Connector line */}
                {index < history.length - 1 && (
                  <span
                    className="absolute left-4 top-4 -ml-px h-full w-0.5 bg-gray-200"
                    aria-hidden="true"
                  />
                )}

                <div className="relative flex space-x-3">
                  {/* Status indicator dot */}
                  <div>
                    <span className="h-8 w-8 rounded-full bg-gray-100 flex items-center justify-center ring-8 ring-white">
                      <div className="h-2 w-2 rounded-full bg-blue-500" />
                    </span>
                  </div>

                  {/* Content */}
                  <div className="flex min-w-0 flex-1 justify-between space-x-4">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      {change.old_status && (
                        <>
                          <StatusBadge status={change.old_status} />
                          <span className="text-gray-400">→</span>
                        </>
                      )}
                      <StatusBadge status={change.new_status} />
                    </div>

                    <p className="text-sm text-gray-500">
                      <span className="font-medium text-gray-700">Source:</span> {change.source}
                    </p>

                    {/* Expandable source details */}
                    {change.source_detail && (
                      <div className="mt-2">
                        <button
                          onClick={() => toggleExpand(index)}
                          className="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800"
                        >
                          {expandedItems.has(index) ? (
                            <>
                              <ChevronUpIcon className="h-4 w-4" />
                              Hide details
                            </>
                          ) : (
                            <>
                              <ChevronDownIcon className="h-4 w-4" />
                              Show details
                            </>
                          )}
                        </button>

                        {expandedItems.has(index) && (
                          <pre className="mt-2 text-xs bg-gray-50 p-2 rounded overflow-x-auto">
                            {JSON.stringify(change.source_detail, null, 2)}
                          </pre>
                        )}
                      </div>
                    )}
                  </div>

                  <div className="whitespace-nowrap text-right text-sm">
                    <time
                      dateTime={change.time}
                      className="text-gray-500"
                      title={formatDate(change.time)}
                    >
                      {formatRelativeTime(change.time)}
                    </time>
                  </div>
                </div>
              </div>
              </div>
            </div>
          </li>
          );
        })}
      </ul>
    </div>
  );
}
