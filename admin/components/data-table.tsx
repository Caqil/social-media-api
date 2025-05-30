// components/data-table-enhanced.tsx
"use client";

import React, { useState, useEffect } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  IconChevronLeft,
  IconChevronRight,
  IconDotsVertical,
  IconSearch,
  IconFilter,
  IconDownload,
  IconRefresh,
  IconSortAscending,
  IconSortDescending,
} from "@tabler/icons-react";
import { TableProps, TableColumn } from "@/types/admin";

export interface DataTableProps extends TableProps {
  title?: string;
  description?: string;
  searchPlaceholder?: string;
  emptyMessage?: string;
  onRefresh?: () => void;
  onExport?: () => void;
  showSearch?: boolean;
  showRefresh?: boolean;
  showExport?: boolean;
}

export function DataTable({
  data,
  columns,
  loading = false,
  pagination,
  title,
  description,
  searchPlaceholder = "Search...",
  emptyMessage = "No data found",
  onPageChange,
  onSort,
  onFilter,
  onRowSelect,
  bulkActions = [],
  onBulkAction,
  onRefresh,
  onExport,
  showSearch = true,
  showRefresh = true,
  showExport = true,
}: DataTableProps) {
  const [selectedRows, setSelectedRows] = useState<string[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [sortColumn, setSortColumn] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<"asc" | "desc">("asc");
  const [filters, setFilters] = useState<Record<string, any>>({});

  // Handle search
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (onFilter) {
        onFilter({ ...filters, search: searchTerm });
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [searchTerm, filters, onFilter]);

  // Handle row selection
  const handleRowSelect = (rowId: string, checked: boolean) => {
    const newSelection = checked
      ? [...selectedRows, rowId]
      : selectedRows.filter((id) => id !== rowId);

    setSelectedRows(newSelection);
    onRowSelect?.(newSelection);
  };

  // Handle select all
  const handleSelectAll = (checked: boolean) => {
    const newSelection = checked ? data.map((row) => row.id) : [];
    setSelectedRows(newSelection);
    onRowSelect?.(newSelection);
  };

  // Handle sorting
  const handleSort = (column: string) => {
    const newDirection =
      column === sortColumn && sortDirection === "asc" ? "desc" : "asc";
    setSortColumn(column);
    setSortDirection(newDirection);
    onSort?.(column, newDirection);
  };

  // Handle bulk action
  const handleBulkAction = (action: string) => {
    if (selectedRows.length > 0) {
      onBulkAction?.(action, selectedRows);
      setSelectedRows([]);
    }
  };

  return (
    <Card>
      {(title || description) && (
        <CardHeader>
          {title && <CardTitle>{title}</CardTitle>}
          {description && (
            <p className="text-sm text-muted-foreground">{description}</p>
          )}
        </CardHeader>
      )}

      <CardContent>
        {/* Toolbar */}
        <div className="flex items-center justify-between gap-4 mb-4">
          <div className="flex items-center gap-2 flex-1">
            {showSearch && (
              <div className="relative max-w-sm">
                <IconSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder={searchPlaceholder}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-9"
                />
              </div>
            )}

            {/* Filter Button */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <IconFilter className="h-4 w-4 mr-2" />
                  Filter
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                {columns
                  .filter((col) => col.filterable)
                  .map((col) => (
                    <DropdownMenuItem key={col.key}>
                      {col.label}
                    </DropdownMenuItem>
                  ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          <div className="flex items-center gap-2">
            {/* Bulk Actions */}
            {selectedRows.length > 0 && bulkActions.length > 0 && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="sm">
                    Actions ({selectedRows.length})
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  {bulkActions.map((action) => (
                    <DropdownMenuItem
                      key={action.action}
                      onClick={() => handleBulkAction(action.action)}
                      className={
                        action.variant === "destructive" ? "text-red-600" : ""
                      }
                    >
                      {action.label}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            )}

            {showRefresh && (
              <Button variant="outline" size="sm" onClick={onRefresh}>
                <IconRefresh className="h-4 w-4" />
              </Button>
            )}

            {showExport && (
              <Button variant="outline" size="sm" onClick={onExport}>
                <IconDownload className="h-4 w-4 mr-2" />
                Export
              </Button>
            )}
          </div>
        </div>

        {/* Table */}
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                {onRowSelect && (
                  <TableHead className="w-12">
                    <Checkbox
                      checked={
                        selectedRows.length === data.length && data.length > 0
                      }
                      onCheckedChange={handleSelectAll}
                    />
                  </TableHead>
                )}
                {columns.map((column) => (
                  <TableHead key={column.key}>
                    <div className="flex items-center gap-2">
                      {column.label}
                      {column.sortable && (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 w-6 p-0"
                          onClick={() => handleSort(column.key)}
                        >
                          {sortColumn === column.key ? (
                            sortDirection === "asc" ? (
                              <IconSortAscending className="h-4 w-4" />
                            ) : (
                              <IconSortDescending className="h-4 w-4" />
                            )
                          ) : (
                            <IconSortAscending className="h-4 w-4 opacity-50" />
                          )}
                        </Button>
                      )}
                    </div>
                  </TableHead>
                ))}
                <TableHead className="w-12">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                // Loading skeleton
                Array.from({ length: 5 }).map((_, index) => (
                  <TableRow key={index}>
                    {onRowSelect && (
                      <TableCell>
                        <Skeleton className="h-4 w-4" />
                      </TableCell>
                    )}
                    {columns.map((column) => (
                      <TableCell key={column.key}>
                        <Skeleton className="h-4 w-20" />
                      </TableCell>
                    ))}
                    <TableCell>
                      <Skeleton className="h-8 w-8" />
                    </TableCell>
                  </TableRow>
                ))
              ) : data.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={columns.length + (onRowSelect ? 2 : 1)}
                    className="text-center py-8"
                  >
                    {emptyMessage}
                  </TableCell>
                </TableRow>
              ) : (
                data.map((row, index) => (
                  <TableRow key={row.id || index}>
                    {onRowSelect && (
                      <TableCell>
                        <Checkbox
                          checked={selectedRows.includes(row.id)}
                          onCheckedChange={(checked) =>
                            handleRowSelect(row.id, checked as boolean)
                          }
                        />
                      </TableCell>
                    )}
                    {columns.map((column) => (
                      <TableCell key={column.key}>
                        {column.render
                          ? column.render(row[column.key], row)
                          : formatCellValue(row[column.key], column.key)}
                      </TableCell>
                    ))}
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 p-0"
                          >
                            <IconDotsVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem>View Details</DropdownMenuItem>
                          <DropdownMenuItem>Edit</DropdownMenuItem>
                          <DropdownMenuItem className="text-red-600">
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>

        {/* Pagination */}
        {pagination && (
          <div className="flex items-center justify-between mt-4">
            <div className="text-sm text-muted-foreground">
              Showing {(pagination.current_page - 1) * pagination.per_page + 1}{" "}
              to{" "}
              {Math.min(
                pagination.current_page * pagination.per_page,
                pagination.total
              )}{" "}
              of {pagination.total} results
            </div>

            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => onPageChange?.(pagination.current_page - 1)}
                disabled={!pagination.has_previous}
              >
                <IconChevronLeft className="h-4 w-4" />
                Previous
              </Button>

              <div className="flex items-center gap-1">
                {Array.from(
                  { length: Math.min(5, pagination.total_pages) },
                  (_, i) => {
                    const page = i + 1;
                    return (
                      <Button
                        key={page}
                        variant={
                          page === pagination.current_page
                            ? "default"
                            : "outline"
                        }
                        size="sm"
                        className="w-8 h-8 p-0"
                        onClick={() => onPageChange?.(page)}
                      >
                        {page}
                      </Button>
                    );
                  }
                )}
              </div>

              <Button
                variant="outline"
                size="sm"
                onClick={() => onPageChange?.(pagination.current_page + 1)}
                disabled={!pagination.has_next}
              >
                Next
                <IconChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// Helper function to format cell values
function formatCellValue(value: any, key: string): React.ReactNode {
  if (value === null || value === undefined) {
    return <span className="text-muted-foreground">-</span>;
  }

  // Handle dates
  if (key.includes("_at") || key.includes("Date")) {
    return new Date(value).toLocaleDateString();
  }

  // Handle booleans
  if (typeof value === "boolean") {
    return (
      <Badge variant={value ? "default" : "secondary"}>
        {value ? "Yes" : "No"}
      </Badge>
    );
  }

  // Handle status fields
  if (key.includes("status")) {
    const statusColors: Record<string, string> = {
      active: "bg-green-100 text-green-800",
      inactive: "bg-gray-100 text-gray-800",
      pending: "bg-yellow-100 text-yellow-800",
      suspended: "bg-red-100 text-red-800",
      resolved: "bg-blue-100 text-blue-800",
    };

    return (
      <Badge
        className={
          statusColors[value?.toLowerCase()] || "bg-gray-100 text-gray-800"
        }
      >
        {value}
      </Badge>
    );
  }

  // Handle arrays
  if (Array.isArray(value)) {
    return value.join(", ");
  }

  return String(value);
}
