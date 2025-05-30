"use client";

import * as React from "react";
import { Area, AreaChart, CartesianGrid, XAxis } from "recharts";

import { useIsMobile } from "@/hooks/use-mobile";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";

export const description = "An interactive area chart";

// Interface for chart data that matches API response
interface ChartDataPoint {
  date: string;
  count?: number;
  value?: number;
  desktop?: number;
  mobile?: number;
  [key: string]: any;
}

interface ChartAreaInteractiveProps {
  data?: ChartDataPoint[];
  title?: string;
  description?: string;
  className?: string;
  height?: number;
  dataKey?: string; // The key to use for the main data (e.g., 'count', 'value')
  showTimeFilter?: boolean;
}

const chartConfig = {
  visitors: {
    label: "Visitors",
  },
  desktop: {
    label: "Desktop",
    color: "var(--primary)",
  },
  mobile: {
    label: "Mobile",
    color: "var(--primary)",
  },
  count: {
    label: "Count",
    color: "var(--primary)",
  },
  value: {
    label: "Value",
    color: "var(--primary)",
  },
} satisfies ChartConfig;

export function ChartAreaInteractive({
  data = [],
  title = "Chart",
  description,
  className,
  height = 250,
  dataKey = "count",
  showTimeFilter = true,
}: ChartAreaInteractiveProps) {
  const isMobile = useIsMobile();
  const [timeRange, setTimeRange] = React.useState("90d");

  React.useEffect(() => {
    if (isMobile) {
      setTimeRange("7d");
    }
  }, [isMobile]);

  // Transform API data to chart format
  const transformedData = React.useMemo(() => {
    if (!data || data.length === 0) return [];

    return data.map((item) => {
      // If data already has the expected format (desktop/mobile), use it
      if (item.desktop !== undefined || item.mobile !== undefined) {
        return item;
      }

      // Otherwise, transform single value data to chart format
      return {
        ...item,
        [dataKey]: item[dataKey] || item.count || item.value || 0,
        // For single metric charts, duplicate the value for visual consistency
        desktop: item[dataKey] || item.count || item.value || 0,
        mobile: 0, // Set to 0 for single metric charts
      };
    });
  }, [data, dataKey]);

  // Filter data based on time range
  const filteredData = React.useMemo(() => {
    if (!transformedData.length) return [];

    if (!showTimeFilter) return transformedData;

    const now = new Date();
    let daysToSubtract = 90;

    if (timeRange === "30d") {
      daysToSubtract = 30;
    } else if (timeRange === "7d") {
      daysToSubtract = 7;
    }

    const startDate = new Date(now);
    startDate.setDate(startDate.getDate() - daysToSubtract);

    return transformedData.filter((item) => {
      const itemDate = new Date(item.date);
      return itemDate >= startDate;
    });
  }, [transformedData, timeRange, showTimeFilter]);

  // Determine if we're showing single metric or dual metric
  const hasDualMetrics = transformedData.some(
    (item) =>
      item.desktop !== undefined && item.mobile !== undefined && item.mobile > 0
  );

  return (
    <Card className={`@container/card ${className || ""}`}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        {description && (
          <CardDescription>
            <span className="hidden @[540px]/card:block">{description}</span>
            <span className="@[540px]/card:hidden">{description}</span>
          </CardDescription>
        )}
        {showTimeFilter && (
          <CardAction>
            <ToggleGroup
              type="single"
              value={timeRange}
              onValueChange={setTimeRange}
              variant="outline"
              className="hidden *:data-[slot=toggle-group-item]:!px-4 @[767px]/card:flex"
            >
              <ToggleGroupItem value="90d">Last 3 months</ToggleGroupItem>
              <ToggleGroupItem value="30d">Last 30 days</ToggleGroupItem>
              <ToggleGroupItem value="7d">Last 7 days</ToggleGroupItem>
            </ToggleGroup>
            <Select value={timeRange} onValueChange={setTimeRange}>
              <SelectTrigger
                className="flex w-40 **:data-[slot=select-value]:block **:data-[slot=select-value]:truncate @[767px]/card:hidden"
                size="sm"
                aria-label="Select a value"
              >
                <SelectValue placeholder="Last 3 months" />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                <SelectItem value="90d" className="rounded-lg">
                  Last 3 months
                </SelectItem>
                <SelectItem value="30d" className="rounded-lg">
                  Last 30 days
                </SelectItem>
                <SelectItem value="7d" className="rounded-lg">
                  Last 7 days
                </SelectItem>
              </SelectContent>
            </Select>
          </CardAction>
        )}
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        {filteredData.length === 0 ? (
          <div className="flex items-center justify-center h-[250px] text-muted-foreground">
            No data available
          </div>
        ) : (
          <ChartContainer
            config={chartConfig}
            className={`aspect-auto h-[${height}px] w-full`}
          >
            <AreaChart data={filteredData}>
              <defs>
                <linearGradient id="fillPrimary" x1="0" y1="0" x2="0" y2="1">
                  <stop
                    offset="5%"
                    stopColor="var(--primary)"
                    stopOpacity={0.8}
                  />
                  <stop
                    offset="95%"
                    stopColor="var(--primary)"
                    stopOpacity={0.1}
                  />
                </linearGradient>
                <linearGradient id="fillSecondary" x1="0" y1="0" x2="0" y2="1">
                  <stop
                    offset="5%"
                    stopColor="var(--primary)"
                    stopOpacity={0.6}
                  />
                  <stop
                    offset="95%"
                    stopColor="var(--primary)"
                    stopOpacity={0.1}
                  />
                </linearGradient>
              </defs>
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                minTickGap={32}
                tickFormatter={(value) => {
                  const date = new Date(value);
                  return date.toLocaleDateString("en-US", {
                    month: "short",
                    day: "numeric",
                  });
                }}
              />
              <ChartTooltip
                cursor={false}
                defaultIndex={
                  isMobile ? -1 : Math.floor(filteredData.length / 2)
                }
                content={
                  <ChartTooltipContent
                    labelFormatter={(value) => {
                      return new Date(value).toLocaleDateString("en-US", {
                        month: "short",
                        day: "numeric",
                        year: "numeric",
                      });
                    }}
                    indicator="dot"
                  />
                }
              />

              {/* Show single metric area for most charts */}
              {!hasDualMetrics && (
                <Area
                  dataKey={dataKey}
                  type="natural"
                  fill="url(#fillPrimary)"
                  stroke="var(--primary)"
                  strokeWidth={2}
                />
              )}

              {/* Show dual metrics if available */}
              {hasDualMetrics && (
                <>
                  <Area
                    dataKey="mobile"
                    type="natural"
                    fill="url(#fillSecondary)"
                    stroke="var(--primary)"
                    strokeWidth={2}
                    stackId="a"
                  />
                  <Area
                    dataKey="desktop"
                    type="natural"
                    fill="url(#fillPrimary)"
                    stroke="var(--primary)"
                    strokeWidth={2}
                    stackId="a"
                  />
                </>
              )}
            </AreaChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
