// components/enhanced-data-table.tsx
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
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  DropdownMenuCheckboxItem,
  DropdownMenuLabel,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  IconChevronLeft,
  IconChevronRight,
  IconChevronsLeft,
  IconChevronsRight,
  IconDotsVertical,
  IconSearch,
  IconFilter,
  IconDownload,
  IconRefresh,
  IconSortAscending,
  IconSortDescending,
  IconColumns,
  IconSettings,
  IconX,
  IconEye,
  IconEyeOff,
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
  showColumnToggle?: boolean;
  showDensityToggle?: boolean;
  customActions?: React.ReactNode;
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
  showColumnToggle = true,
  showDensityToggle = true,
  customActions,
}: DataTableProps) {
  const [selectedRows, setSelectedRows] = useState<string[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [sortColumn, setSortColumn] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<"asc" | "desc">("asc");
  const [filters, setFilters] = useState<Record<string, any>>({});
  const [visibleColumns, setVisibleColumns] = useState<string[]>(
    columns.map((col) => col.key)
  );
  const [density, setDensity] = useState<
    "compact" | "comfortable" | "spacious"
  >("comfortable");
  const [showFilters, setShowFilters] = useState(false);

  // Handle search with debounce
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

  // Handle column visibility toggle
  const toggleColumnVisibility = (columnKey: string) => {
    setVisibleColumns((prev) =>
      prev.includes(columnKey)
        ? prev.filter((key) => key !== columnKey)
        : [...prev, columnKey]
    );
  };

  // Handle filter change
  const handleFilterChange = (key: string, value: any) => {
    const newFilters = { ...filters };
    if (value === "" || value === null || value === undefined) {
      delete newFilters[key];
    } else {
      newFilters[key] = value;
    }
    setFilters(newFilters);
  };

  // Clear all filters
  const clearFilters = () => {
    setFilters({});
    setSearchTerm("");
  };

  // Get visible columns
  const getVisibleColumns = () => {
    return columns.filter((col) => visibleColumns.includes(col.key));
  };

  // Get filterable columns
  const getFilterableColumns = () => {
    return columns.filter((col) => col.filterable);
  };

  // Get density classes
  const getDensityClasses = () => {
    switch (density) {
      case "compact":
        return "text-xs";
      case "spacious":
        return "text-base py-4";
      default:
        return "text-sm py-2";
    }
  };

  // Active filters count
  const activeFiltersCount = Object.keys(filters).length + (searchTerm ? 1 : 0);

  return (
    <Card className="w-full">
      {(title || description) && (
        <CardHeader className="pb-4">
          <div className="flex items-center justify-between">
            <div>
              {title && <CardTitle className="text-xl">{title}</CardTitle>}
              {description && (
                <p className="text-sm text-muted-foreground mt-1">
                  {description}
                </p>
              )}
            </div>
            {customActions}
          </div>
        </CardHeader>
      )}

      <CardContent className="p-0">
        {/* Toolbar */}
        <div className="p-4 border-b bg-muted/20">
          <div className="flex items-center justify-between gap-4 mb-4">
            <div className="flex items-center gap-2 flex-1">
              {/* Search */}
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

              {/* Filters Toggle */}
              <Button
                variant={showFilters ? "default" : "outline"}
                size="sm"
                onClick={() => setShowFilters(!showFilters)}
              >
                <IconFilter className="h-4 w-4 mr-2" />
                Filters
                {activeFiltersCount > 0 && (
                  <Badge variant="secondary" className="ml-2">
                    {activeFiltersCount}
                  </Badge>
                )}
              </Button>

              {/* Clear Filters */}
              {activeFiltersCount > 0 && (
                <Button variant="ghost" size="sm" onClick={clearFilters}>
                  <IconX className="h-4 w-4 mr-2" />
                  Clear
                </Button>
              )}
            </div>

            <div className="flex items-center gap-2">
              {/* Bulk Actions */}
              {selectedRows.length > 0 && bulkActions.length > 0 && (
                <>
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
                            action.variant === "destructive"
                              ? "text-red-600"
                              : ""
                          }
                        >
                          {action.label}
                        </DropdownMenuItem>
                      ))}
                    </DropdownMenuContent>
                  </DropdownMenu>
                  <Separator orientation="vertical" className="h-4" />
                </>
              )}

              {/* Column Toggle */}
              {showColumnToggle && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="sm">
                      <IconColumns className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-48">
                    <DropdownMenuLabel>Toggle Columns</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    {columns.map((column) => (
                      <DropdownMenuCheckboxItem
                        key={column.key}
                        checked={visibleColumns.includes(column.key)}
                        onCheckedChange={() =>
                          toggleColumnVisibility(column.key)
                        }
                      >
                        {column.label}
                      </DropdownMenuCheckboxItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
              )}

              {/* Density Toggle */}
              {showDensityToggle && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="sm">
                      <IconSettings className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuLabel>Display Density</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={() => setDensity("compact")}>
                      Compact
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => setDensity("comfortable")}>
                      Comfortable
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => setDensity("spacious")}>
                      Spacious
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              )}

              {/* Refresh */}
              {showRefresh && (
                <Button variant="outline" size="sm" onClick={onRefresh}>
                  <IconRefresh className="h-4 w-4" />
                </Button>
              )}

              {/* Export */}
              {showExport && (
                <Button variant="outline" size="sm" onClick={onExport}>
                  <IconDownload className="h-4 w-4 mr-2" />
                  Export
                </Button>
              )}
            </div>
          </div>

          {/* Advanced Filters */}
          {showFilters && (
            <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4 p-4 bg-background rounded-lg border">
              {getFilterableColumns().map((column) => (
                <div key={column.key} className="space-y-2">
                  <label className="text-sm font-medium">{column.label}</label>
                  <Select
                    value={filters[column.key] || ""}
                    onValueChange={(value) =>
                      handleFilterChange(column.key, value)
                    }
                  >
                    <SelectTrigger className="h-8">
                      <SelectValue
                        placeholder={`Filter by ${column.label.toLowerCase()}`}
                      />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">All</SelectItem>
                      {/* This would be populated with actual filter options */}
                      <SelectItem value="active">Active</SelectItem>
                      <SelectItem value="inactive">Inactive</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Table */}
        <div className="relative overflow-auto">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
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
                {getVisibleColumns().map((column) => (
                  <TableHead key={column.key} className={column.width}>
                    <div className="flex items-center gap-2">
                      <span
                        className={
                          column.align === "center"
                            ? "text-center"
                            : column.align === "right"
                            ? "text-right"
                            : ""
                        }
                      >
                        {column.label}
                      </span>
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
                <TableHead className="w-12"></TableHead>
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
                    {getVisibleColumns().map((column) => (
                      <TableCell key={column.key}>
                        <Skeleton className="h-4 w-full max-w-32" />
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
                    colSpan={getVisibleColumns().length + (onRowSelect ? 2 : 1)}
                    className="text-center py-12"
                  >
                    <div className="flex flex-col items-center gap-2">
                      <IconEyeOff className="h-8 w-8 text-muted-foreground" />
                      <p className="text-muted-foreground">{emptyMessage}</p>
                    </div>
                  </TableCell>
                </TableRow>
              ) : (
                data.map((row, index) => (
                  <TableRow
                    key={row.id || index}
                    className={`${getDensityClasses()} hover:bg-muted/50 transition-colors`}
                  >
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
                    {getVisibleColumns().map((column) => (
                      <TableCell
                        key={column.key}
                        className={
                          column.align === "center"
                            ? "text-center"
                            : column.align === "right"
                            ? "text-right"
                            : ""
                        }
                      >
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
                          <DropdownMenuItem>
                            <IconEye className="h-4 w-4 mr-2" />
                            View Details
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
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
          <div className="flex items-center justify-between p-4 border-t bg-muted/20">
            <div className="text-sm text-muted-foreground">
              Showing{" "}
              <span className="font-medium">
                {(pagination.current_page - 1) * pagination.per_page + 1}
              </span>{" "}
              to{" "}
              <span className="font-medium">
                {Math.min(
                  pagination.current_page * pagination.per_page,
                  pagination.total
                )}
              </span>{" "}
              of <span className="font-medium">{pagination.total}</span> results
            </div>

            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => onPageChange?.(1)}
                disabled={!pagination.has_previous}
              >
                <IconChevronsLeft className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => onPageChange?.(pagination.current_page - 1)}
                disabled={!pagination.has_previous}
              >
                <IconChevronLeft className="h-4 w-4" />
              </Button>

              <div className="flex items-center gap-1">
                {Array.from(
                  { length: Math.min(5, pagination.total_pages) },
                  (_, i) => {
                    const page = Math.max(
                      1,
                      Math.min(
                        pagination.current_page - 2 + i,
                        pagination.total_pages
                      )
                    );
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
                <IconChevronRight className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => onPageChange?.(pagination.total_pages)}
                disabled={!pagination.has_next}
              >
                <IconChevronsRight className="h-4 w-4" />
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
