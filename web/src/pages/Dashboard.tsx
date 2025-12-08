import { useState, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { PlusIcon } from '@heroicons/react/24/outline';
import Button from '@/components/common/Button';
import LineList from '@/components/lines/LineList';
import LabelFilter from '@/components/labels/LabelFilter';
import { useLines, useLabels } from '@/hooks/useLines';
import type { Label } from '@/api/types';

export default function Dashboard() {
  const [selectedLabels, setSelectedLabels] = useState<Label[]>([]);
  const { data: allLines, isLoading, error } = useLines();
  const { data: availableLabels } = useLabels();

  // Filter lines by selected labels
  const filteredLines = useMemo(() => {
    if (!allLines) return [];
    if (selectedLabels.length === 0) return allLines;

    return allLines.filter((line) =>
      line.labels?.some((lineLabel) =>
        selectedLabels.some((selectedLabel) => selectedLabel.id === lineLabel.id)
      )
    );
  }, [allLines, selectedLabels]);

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-gray-600">
            Manage and monitor all production lines
          </p>
        </div>
        <Link to="/lines/new">
          <Button variant="primary">
            <PlusIcon className="h-5 w-5 mr-2" />
            Create New Line
          </Button>
        </Link>
      </div>

      {/* Label filter */}
      {availableLabels && availableLabels.length > 0 && (
        <div className="max-w-xs">
          <LabelFilter
            selectedLabels={selectedLabels}
            onLabelsChange={setSelectedLabels}
            availableLabels={availableLabels}
          />
        </div>
      )}

      {/* Lines list */}
      <LineList lines={filteredLines} isLoading={isLoading} error={error} />
    </div>
  );
}
