import { Fragment, useState } from 'react';
import { Dialog, DialogPanel, DialogTitle, Transition, TransitionChild, Combobox, ComboboxInput, ComboboxButton, ComboboxOptions, ComboboxOption } from '@headlessui/react';
import { XMarkIcon, ChevronUpDownIcon } from '@heroicons/react/24/outline';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import type { CreateLineScheduleExceptionRequest, CreateScheduleDayInput, ProductionLine } from '@/api/types';
import WeeklyScheduleEditor from './WeeklyScheduleEditor';
import { createDefaultDays } from './scheduleUtils';

const exceptionSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  start_date: z.string().min(1, 'Start date is required'),
  end_date: z.string().min(1, 'End date is required'),
}).refine((data) => data.start_date <= data.end_date, {
  message: 'End date must be on or after start date',
  path: ['end_date'],
});

type ExceptionFormData = z.infer<typeof exceptionSchema>;

interface LineExceptionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateLineScheduleExceptionRequest) => void;
  availableLines: ProductionLine[];
  isLoading?: boolean;
}

export default function LineExceptionModal({
  isOpen,
  onClose,
  onSubmit,
  availableLines,
  isLoading = false,
}: LineExceptionModalProps) {
  const [days, setDays] = useState<CreateScheduleDayInput[]>(createDefaultDays());
  const [selectedLines, setSelectedLines] = useState<ProductionLine[]>([]);
  const [query, setQuery] = useState('');

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setError,
  } = useForm<ExceptionFormData>({
    resolver: zodResolver(exceptionSchema),
    defaultValues: {
      name: '',
      start_date: '',
      end_date: '',
    },
  });

  const filteredLines = query === ''
    ? availableLines.filter(line => !selectedLines.some(s => s.id === line.id))
    : availableLines.filter(
        line =>
          !selectedLines.some(s => s.id === line.id) &&
          (line.name.toLowerCase().includes(query.toLowerCase()) ||
           line.code.toLowerCase().includes(query.toLowerCase()))
      );

  const onFormSubmit = (data: ExceptionFormData) => {
    if (selectedLines.length === 0) {
      setError('name', { message: 'At least one line must be selected' });
      return;
    }

    const request: CreateLineScheduleExceptionRequest = {
      name: data.name,
      start_date: data.start_date,
      end_date: data.end_date,
      line_ids: selectedLines.map(l => l.id),
      days,
    };
    onSubmit(request);
    reset();
    setDays(createDefaultDays());
    setSelectedLines([]);
  };

  const handleClose = () => {
    reset();
    setDays(createDefaultDays());
    setSelectedLines([]);
    setQuery('');
    onClose();
  };

  const removeLine = (lineId: string) => {
    setSelectedLines(selectedLines.filter(l => l.id !== lineId));
  };

  return (
    <Transition show={isOpen} as={Fragment}>
      <Dialog as="div" className="relative z-50" onClose={handleClose}>
        <TransitionChild
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
        </TransitionChild>

        <div className="fixed inset-0 z-10 overflow-y-auto">
          <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
            <TransitionChild
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              enterTo="opacity-100 translate-y-0 sm:scale-100"
              leave="ease-in duration-200"
              leaveFrom="opacity-100 translate-y-0 sm:scale-100"
              leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            >
              <DialogPanel className="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-4xl sm:p-6">
                <div className="absolute right-0 top-0 pr-4 pt-4">
                  <button
                    type="button"
                    className="rounded-md bg-white text-gray-400 hover:text-gray-500"
                    onClick={handleClose}
                  >
                    <span className="sr-only">Close</span>
                    <XMarkIcon className="h-6 w-6" />
                  </button>
                </div>

                <DialogTitle as="h3" className="text-lg font-semibold text-gray-900 mb-4">
                  Add Line-Specific Exception
                </DialogTitle>

                <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-6">
                  {/* Basic Info */}
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="md:col-span-3">
                      <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                        Exception Name *
                      </label>
                      <input
                        id="name"
                        type="text"
                        {...register('name')}
                        className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="e.g., Line Maintenance"
                      />
                      {errors.name && (
                        <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
                      )}
                    </div>

                    <div>
                      <label htmlFor="start_date" className="block text-sm font-medium text-gray-700 mb-1">
                        Start Date *
                      </label>
                      <input
                        id="start_date"
                        type="date"
                        {...register('start_date')}
                        className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                      {errors.start_date && (
                        <p className="mt-1 text-sm text-red-600">{errors.start_date.message}</p>
                      )}
                    </div>

                    <div>
                      <label htmlFor="end_date" className="block text-sm font-medium text-gray-700 mb-1">
                        End Date *
                      </label>
                      <input
                        id="end_date"
                        type="date"
                        {...register('end_date')}
                        className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                      {errors.end_date && (
                        <p className="mt-1 text-sm text-red-600">{errors.end_date.message}</p>
                      )}
                    </div>
                  </div>

                  {/* Line Selection */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Select Lines *
                    </label>
                    <Combobox value={null} onChange={(line: ProductionLine | null) => {
                      if (line) {
                        setSelectedLines([...selectedLines, line]);
                        setQuery('');
                      }
                    }}>
                      <div className="relative">
                        <ComboboxInput
                          className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
                          placeholder="Search and select lines..."
                          onChange={(event) => setQuery(event.target.value)}
                          displayValue={() => ''}
                        />
                        <ComboboxButton className="absolute inset-y-0 right-0 flex items-center pr-2">
                          <ChevronUpDownIcon className="h-5 w-5 text-gray-400" />
                        </ComboboxButton>
                        <ComboboxOptions className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                          {filteredLines.length === 0 ? (
                            <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                              No lines found.
                            </div>
                          ) : (
                            filteredLines.map((line) => (
                              <ComboboxOption
                                key={line.id}
                                value={line}
                                className={({ focus }) =>
                                  `relative cursor-default select-none py-2 pl-10 pr-4 ${
                                    focus ? 'bg-blue-600 text-white' : 'text-gray-900'
                                  }`
                                }
                              >
                                <span className="block truncate font-normal">
                                  {line.name} ({line.code})
                                </span>
                              </ComboboxOption>
                            ))
                          )}
                        </ComboboxOptions>
                      </div>
                    </Combobox>

                    {/* Selected Lines */}
                    {selectedLines.length > 0 && (
                      <div className="mt-2 flex flex-wrap gap-2">
                        {selectedLines.map((line) => (
                          <span
                            key={line.id}
                            className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-sm bg-blue-100 text-blue-800"
                          >
                            {line.name}
                            <button
                              type="button"
                              onClick={() => removeLine(line.id)}
                              className="hover:text-blue-600"
                            >
                              <XMarkIcon className="h-4 w-4" />
                            </button>
                          </span>
                        ))}
                      </div>
                    )}
                  </div>

                  {/* Weekly Schedule Override */}
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-2">
                      Override Schedule (for this date range)
                    </h4>
                    <div className="border border-gray-200 rounded-lg overflow-hidden">
                      <WeeklyScheduleEditor days={days} onChange={setDays} disabled={isLoading} />
                    </div>
                  </div>

                  <div className="flex justify-end gap-3 pt-4">
                    <button
                      type="button"
                      onClick={handleClose}
                      disabled={isLoading}
                      className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 disabled:opacity-50"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={isLoading || selectedLines.length === 0}
                      className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md shadow-sm hover:bg-blue-700 disabled:opacity-50"
                    >
                      {isLoading ? 'Saving...' : 'Add Line Exception'}
                    </button>
                  </div>
                </form>
              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </Dialog>
    </Transition>
  );
}
