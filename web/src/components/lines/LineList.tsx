import { useLines } from '@/hooks/useLines';
import Loading from '@/components/common/Loading';
import LineCard from './LineCard';

export default function LineList() {
  const { data: lines, isLoading, error } = useLines();

  if (isLoading) {
    return <Loading message="Loading production lines..." />;
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6">
        <h3 className="text-red-800 font-semibold mb-2">Error Loading Lines</h3>
        <p className="text-red-600">
          {error instanceof Error ? error.message : 'An unexpected error occurred'}
        </p>
      </div>
    );
  }

  if (!lines || lines.length === 0) {
    return (
      <div className="bg-gray-50 border border-gray-200 rounded-lg p-12 text-center">
        <p className="text-gray-600 text-lg mb-4">No production lines found</p>
        <p className="text-gray-500 text-sm">
          Create your first production line to get started
        </p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {lines.map((line) => (
        <LineCard key={line.id} line={line} />
      ))}
    </div>
  );
}
