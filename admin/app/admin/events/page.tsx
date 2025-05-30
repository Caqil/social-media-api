// app/admin/events/page.tsx
"use client";

import { useEffect, useState } from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { SiteHeader } from "@/components/site-header";
import { DataTable } from "@/components/data-table";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { Event, EventAttendee, TableColumn } from "@/types/admin";
import {
  IconCalendarEvent,
  IconEye,
  IconTrash,
  IconCheck,
  IconX,
  IconUsers,
  IconMapPin,
  IconClock,
  IconDollarSign,
  IconGlobe,
  IconLock,
  IconEyeOff,
  IconCalendar,
  IconStar,
} from "@tabler/icons-react";

function EventsPage() {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<any>(null);
  const [filters, setFilters] = useState<any>({ page: 1, limit: 20 });

  // Dialog states
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [showEventDetails, setShowEventDetails] = useState(false);
  const [showAttendeesDialog, setShowAttendeesDialog] = useState(false);
  const [showActionDialog, setShowActionDialog] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [actionType, setActionType] = useState<string>("");
  const [bulkAction, setBulkAction] = useState<string>("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [actionReason, setActionReason] = useState("");

  // Attendees data
  const [attendees, setAttendees] = useState<EventAttendee[]>([]);
  const [attendeesLoading, setAttendeesLoading] = useState(false);
  const [attendeesPagination, setAttendeesPagination] = useState<any>(null);

  // Fetch events
  const fetchEvents = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getEvents(filters);
      setEvents(response.data.data);
      setPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch events:", error);
      setError(error.response?.data?.message || "Failed to load events");
    } finally {
      setLoading(false);
    }
  };

  // Fetch event attendees
  const fetchEventAttendees = async (eventId: string) => {
    try {
      setAttendeesLoading(true);
      const response = await apiClient.getEventAttendees(eventId);
      setAttendees(response.data.data);
      setAttendeesPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch event attendees:", error);
    } finally {
      setAttendeesLoading(false);
    }
  };

  useEffect(() => {
    fetchEvents();
  }, [filters]);

  // Handle page change
  const handlePageChange = (page: number) => {
    setFilters((prev: any) => ({ ...prev, page }));
  };

  // Handle filtering
  const handleFilter = (newFilters: Record<string, any>) => {
    setFilters((prev: any) => ({
      ...prev,
      ...newFilters,
      page: 1,
    }));
  };

  // Handle row selection
  const handleRowSelect = (selectedRows: string[]) => {
    setSelectedIds(selectedRows);
  };

  // Handle bulk actions
  const handleBulkAction = (action: string, ids: string[]) => {
    setBulkAction(action);
    setSelectedIds(ids);
    setShowBulkDialog(true);
  };

  // Execute bulk action
  const executeBulkAction = async () => {
    try {
      await apiClient.bulkEventAction({
        event_ids: selectedIds,
        action: bulkAction,
        reason: actionReason,
      });

      setShowBulkDialog(false);
      setActionReason("");
      fetchEvents();
    } catch (error: any) {
      console.error("Bulk action failed:", error);
      setError(error.response?.data?.message || "Bulk action failed");
    }
  };

  // Handle individual event action
  const handleEventAction = (event: Event, action: string) => {
    setSelectedEvent(event);
    setActionType(action);
    setShowActionDialog(true);
  };

  // Execute event action
  const executeEventAction = async () => {
    if (!selectedEvent) return;

    try {
      switch (actionType) {
        case "publish":
          await apiClient.updateEventStatus(selectedEvent.id, {
            status: "published",
            reason: actionReason,
          });
          break;
        case "cancel":
          await apiClient.updateEventStatus(selectedEvent.id, {
            status: "cancelled",
            reason: actionReason,
          });
          break;
        case "complete":
          await apiClient.updateEventStatus(selectedEvent.id, {
            status: "completed",
            reason: actionReason,
          });
          break;
        case "delete":
          await apiClient.deleteEvent(selectedEvent.id, actionReason);
          break;
      }

      setShowActionDialog(false);
      setActionReason("");
      fetchEvents();
    } catch (error: any) {
      console.error("Event action failed:", error);
      setError(error.response?.data?.message || "Event action failed");
    }
  };

  // View event details
  const viewEventDetails = async (event: Event) => {
    try {
      const response = await apiClient.getEvent(event.id);
      setSelectedEvent(response.data);
      setShowEventDetails(true);
    } catch (error) {
      console.error("Failed to fetch event details:", error);
    }
  };

  // View event attendees
  const viewEventAttendees = (event: Event) => {
    setSelectedEvent(event);
    fetchEventAttendees(event.id);
    setShowAttendeesDialog(true);
  };

  // Get event status color
  const getEventStatusColor = (status: string) => {
    switch (status) {
      case "published":
        return "bg-green-100 text-green-800";
      case "draft":
        return "bg-gray-100 text-gray-800";
      case "cancelled":
        return "bg-red-100 text-red-800";
      case "completed":
        return "bg-blue-100 text-blue-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  // Get visibility icon
  const getVisibilityIcon = (visibility: string) => {
    switch (visibility) {
      case "public":
        return <IconGlobe className="h-4 w-4" />;
      case "friends":
        return <IconUsers className="h-4 w-4" />;
      case "private":
        return <IconLock className="h-4 w-4" />;
      default:
        return <IconGlobe className="h-4 w-4" />;
    }
  };

  // Check if event is upcoming
  const isUpcoming = (startDate: string) => {
    return new Date(startDate) > new Date();
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "event",
      label: "Event",
      render: (_, event: Event) => (
        <div className="max-w-md">
          <div className="flex items-center gap-2 mb-2">
            <IconCalendarEvent className="h-4 w-4" />
            {isUpcoming(event.start_date) && (
              <Badge variant="outline" className="text-xs">
                Upcoming
              </Badge>
            )}
          </div>
          <h4 className="font-medium mb-1">{event.title}</h4>
          <p className="text-sm text-muted-foreground line-clamp-2 mb-2">
            {event.description}
          </p>
          {event.creator && (
            <div className="flex items-center gap-2">
              <Avatar className="h-5 w-5">
                <AvatarImage src={event.creator.profile_picture} />
                <AvatarFallback className="text-xs">
                  {event.creator.username?.[0]?.toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span className="text-xs text-muted-foreground">
                by {event.creator.username}
              </span>
            </div>
          )}
        </div>
      ),
    },
    {
      key: "category",
      label: "Category",
      filterable: true,
      render: (value: string) => <Badge variant="secondary">{value}</Badge>,
    },
    {
      key: "date_info",
      label: "Date & Time",
      render: (_, event: Event) => (
        <div className="text-sm space-y-1">
          <div className="flex items-center gap-1">
            <IconCalendar className="h-3 w-3" />
            <span>{new Date(event.start_date).toLocaleDateString()}</span>
          </div>
          <div className="flex items-center gap-1">
            <IconClock className="h-3 w-3" />
            <span>{new Date(event.start_date).toLocaleTimeString()}</span>
          </div>
          {event.end_date && (
            <div className="text-xs text-muted-foreground">
              Until {new Date(event.end_date).toLocaleDateString()}
            </div>
          )}
        </div>
      ),
    },
    {
      key: "location",
      label: "Location",
      render: (_, event: Event) =>
        event.location ? (
          <div className="text-sm">
            <div className="flex items-center gap-1 mb-1">
              <IconMapPin className="h-3 w-3" />
              <span>{event.location}</span>
            </div>
            {event.address && (
              <p className="text-xs text-muted-foreground">{event.address}</p>
            )}
          </div>
        ) : (
          <span className="text-muted-foreground">Online</span>
        ),
    },
    {
      key: "attendees",
      label: "Attendees",
      render: (_, event: Event) => (
        <div
          className="cursor-pointer hover:text-primary"
          onClick={() => viewEventAttendees(event)}
        >
          <div className="flex items-center gap-1 mb-1">
            <IconUsers className="h-3 w-3" />
            <span className="font-medium">{event.attendees_count || 0}</span>
          </div>
          <div className="flex items-center gap-1">
            <IconStar className="h-3 w-3" />
            <span className="text-xs">
              {event.interested_count || 0} interested
            </span>
          </div>
          {event.capacity && (
            <div className="text-xs text-muted-foreground">
              / {event.capacity} max
            </div>
          )}
        </div>
      ),
    },
    {
      key: "price",
      label: "Price",
      render: (_, event: Event) =>
        event.price ? (
          <div className="flex items-center gap-1">
            <IconDollarSign className="h-3 w-3" />
            <span>
              {event.price} {event.currency || "USD"}
            </span>
          </div>
        ) : (
          <Badge variant="outline">Free</Badge>
        ),
    },
    {
      key: "visibility",
      label: "Visibility",
      filterable: true,
      render: (value: string) => (
        <div className="flex items-center gap-1">
          {getVisibilityIcon(value)}
          <span className="capitalize text-sm">{value}</span>
        </div>
      ),
    },
    {
      key: "status",
      label: "Status",
      filterable: true,
      render: (value: string) => (
        <Badge className={getEventStatusColor(value)}>
          {value === "published" && <IconCheck className="h-3 w-3 mr-1" />}
          {value === "draft" && <IconEyeOff className="h-3 w-3 mr-1" />}
          {value === "cancelled" && <IconX className="h-3 w-3 mr-1" />}
          {value === "completed" && <IconCheck className="h-3 w-3 mr-1" />}
          <span className="capitalize">{value}</span>
        </Badge>
      ),
    },
    {
      key: "created_at",
      label: "Created",
      sortable: true,
      render: (value: string) => (
        <div className="text-sm">{new Date(value).toLocaleDateString()}</div>
      ),
    },
    {
      key: "actions",
      label: "Actions",
      render: (_, event: Event) => (
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewEventDetails(event)}
          >
            <IconEye className="h-3 w-3" />
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewEventAttendees(event)}
          >
            <IconUsers className="h-3 w-3" />
          </Button>
          {event.status === "draft" && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleEventAction(event, "publish")}
            >
              <IconCheck className="h-3 w-3" />
            </Button>
          )}
          {event.status === "published" && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleEventAction(event, "cancel")}
            >
              <IconX className="h-3 w-3" />
            </Button>
          )}
          <Button
            size="sm"
            variant="destructive"
            onClick={() => handleEventAction(event, "delete")}
          >
            <IconTrash className="h-3 w-3" />
          </Button>
        </div>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    { label: "Publish Events", action: "publish", variant: "default" as const },
    {
      label: "Cancel Events",
      action: "cancel",
      variant: "destructive" as const,
    },
    {
      label: "Delete Events",
      action: "delete",
      variant: "destructive" as const,
    },
  ];

  if (error) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Alert variant="destructive" className="max-w-md">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <SidebarProvider
      style={
        {
          "--sidebar-width": "calc(var(--spacing) * 72)",
          "--header-height": "calc(var(--spacing) * 12)",
        } as React.CSSProperties
      }
    >
      <AppSidebar variant="inset" />
      <SidebarInset>
        <SiteHeader />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">Events Management</h1>
              <p className="text-muted-foreground">
                Manage community events and activities
              </p>
            </div>
          </div>

          <DataTable
            title="Events"
            description={`Manage ${pagination?.total || 0} events`}
            data={events}
            columns={columns}
            loading={loading}
            pagination={pagination}
            searchPlaceholder="Search events by title or description..."
            onPageChange={handlePageChange}
            onFilter={handleFilter}
            onRowSelect={handleRowSelect}
            bulkActions={bulkActions}
            onBulkAction={handleBulkAction}
            onRefresh={fetchEvents}
            onExport={() => console.log("Export events")}
          />
        </div>
      </SidebarInset>

      {/* Event Details Dialog */}
      <Dialog open={showEventDetails} onOpenChange={setShowEventDetails}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Event Details</DialogTitle>
          </DialogHeader>

          {selectedEvent && (
            <Tabs defaultValue="details" className="w-full">
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="details">Details</TabsTrigger>
                <TabsTrigger value="settings">Settings</TabsTrigger>
                <TabsTrigger value="stats">Statistics</TabsTrigger>
              </TabsList>

              <TabsContent value="details" className="space-y-4">
                <Card>
                  <CardHeader>
                    <div className="flex items-center gap-4">
                      {selectedEvent.image && (
                        <img
                          src={selectedEvent.image}
                          alt={selectedEvent.title}
                          className="w-16 h-16 rounded-lg object-cover"
                        />
                      )}
                      <div className="flex-1">
                        <CardTitle className="flex items-center gap-2">
                          {selectedEvent.title}
                          <Badge
                            className={getEventStatusColor(
                              selectedEvent.status
                            )}
                          >
                            {selectedEvent.status}
                          </Badge>
                        </CardTitle>
                        <CardDescription>
                          {selectedEvent.description}
                        </CardDescription>
                        <div className="flex items-center gap-4 mt-2 text-sm">
                          <div className="flex items-center gap-1">
                            <IconCalendar className="h-4 w-4" />
                            {new Date(
                              selectedEvent.start_date
                            ).toLocaleDateString()}
                          </div>
                          <div className="flex items-center gap-1">
                            <IconClock className="h-4 w-4" />
                            {new Date(
                              selectedEvent.start_date
                            ).toLocaleTimeString()}
                          </div>
                          {selectedEvent.location && (
                            <div className="flex items-center gap-1">
                              <IconMapPin className="h-4 w-4" />
                              {selectedEvent.location}
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div>
                      <h4 className="font-medium mb-2">Category</h4>
                      <Badge>{selectedEvent.category}</Badge>
                    </div>

                    {selectedEvent.tags && selectedEvent.tags.length > 0 && (
                      <div>
                        <h4 className="font-medium mb-2">Tags</h4>
                        <div className="flex flex-wrap gap-2">
                          {selectedEvent.tags.map((tag, index) => (
                            <Badge key={index} variant="outline">
                              #{tag}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}

                    {selectedEvent.address && (
                      <div>
                        <h4 className="font-medium mb-2">Full Address</h4>
                        <p className="text-sm">{selectedEvent.address}</p>
                      </div>
                    )}

                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <h4 className="font-medium mb-2">Price</h4>
                        <p className="text-sm">
                          {selectedEvent.price ? (
                            <>
                              {selectedEvent.price}{" "}
                              {selectedEvent.currency || "USD"}
                            </>
                          ) : (
                            "Free"
                          )}
                        </p>
                      </div>
                      <div>
                        <h4 className="font-medium mb-2">Capacity</h4>
                        <p className="text-sm">
                          {selectedEvent.capacity
                            ? selectedEvent.capacity
                            : "Unlimited"}
                        </p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="settings" className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle>Event Settings</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="text-sm font-medium">
                          Require Approval
                        </label>
                        <p className="text-sm text-muted-foreground">
                          {selectedEvent.settings?.require_approval
                            ? "Yes"
                            : "No"}
                        </p>
                      </div>
                      <div>
                        <label className="text-sm font-medium">
                          Allow Guests
                        </label>
                        <p className="text-sm text-muted-foreground">
                          {selectedEvent.settings?.allow_guests ? "Yes" : "No"}
                        </p>
                      </div>
                      <div>
                        <label className="text-sm font-medium">
                          Show Attendee List
                        </label>
                        <p className="text-sm text-muted-foreground">
                          {selectedEvent.settings?.show_attendee_list
                            ? "Yes"
                            : "No"}
                        </p>
                      </div>
                      <div>
                        <label className="text-sm font-medium">
                          Send Reminders
                        </label>
                        <p className="text-sm text-muted-foreground">
                          {selectedEvent.settings?.send_reminders
                            ? "Yes"
                            : "No"}
                        </p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="stats" className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">Attendees</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {selectedEvent.attendees_count?.toLocaleString() || 0}
                      </div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">Interested</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {selectedEvent.interested_count?.toLocaleString() || 0}
                      </div>
                    </CardContent>
                  </Card>
                </div>

                <div className="text-sm text-muted-foreground space-y-1">
                  <p>
                    Created:{" "}
                    {new Date(selectedEvent.created_at).toLocaleDateString()}
                  </p>
                  <p>
                    Last Updated:{" "}
                    {new Date(selectedEvent.updated_at).toLocaleDateString()}
                  </p>
                  <p>Status: {selectedEvent.status}</p>
                  <p>Visibility: {selectedEvent.visibility}</p>
                </div>
              </TabsContent>
            </Tabs>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowEventDetails(false)}
            >
              Close
            </Button>
            {selectedEvent && (
              <>
                <Button
                  variant="outline"
                  onClick={() => viewEventAttendees(selectedEvent)}
                >
                  View Attendees
                </Button>
                {selectedEvent.status === "draft" && (
                  <Button
                    onClick={() => {
                      setShowEventDetails(false);
                      handleEventAction(selectedEvent, "publish");
                    }}
                  >
                    Publish Event
                  </Button>
                )}
                {selectedEvent.status === "published" && (
                  <Button
                    variant="outline"
                    onClick={() => {
                      setShowEventDetails(false);
                      handleEventAction(selectedEvent, "cancel");
                    }}
                  >
                    Cancel Event
                  </Button>
                )}
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Event Attendees Dialog */}
      <Dialog open={showAttendeesDialog} onOpenChange={setShowAttendeesDialog}>
        <DialogContent className="max-w-4xl max-h-[80vh]">
          <DialogHeader>
            <DialogTitle>
              {selectedEvent?.title} - Attendees (
              {selectedEvent?.attendees_count})
            </DialogTitle>
          </DialogHeader>

          <div className="space-y-4 max-h-96 overflow-y-auto">
            {attendeesLoading ? (
              <div className="text-center py-8">Loading attendees...</div>
            ) : attendees.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No attendees found
              </div>
            ) : (
              <div className="space-y-2">
                {attendees.map((attendee) => (
                  <div
                    key={attendee.id}
                    className="flex items-center justify-between p-3 border rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      {attendee.user && (
                        <>
                          <Avatar>
                            <AvatarImage src={attendee.user.profile_picture} />
                            <AvatarFallback>
                              {attendee.user.username?.[0]?.toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div>
                            <p className="font-medium">
                              {attendee.user.username}
                            </p>
                            <p className="text-sm text-muted-foreground">
                              Responded{" "}
                              {new Date(
                                attendee.response_date
                              ).toLocaleDateString()}
                            </p>
                          </div>
                        </>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge
                        variant={
                          attendee.status === "attending"
                            ? "default"
                            : attendee.status === "interested"
                            ? "secondary"
                            : "outline"
                        }
                      >
                        <span className="capitalize">{attendee.status}</span>
                      </Badge>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowAttendeesDialog(false)}
            >
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Action Dialog */}
      <Dialog open={showActionDialog} onOpenChange={setShowActionDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {actionType === "publish" && "Publish Event"}
              {actionType === "cancel" && "Cancel Event"}
              {actionType === "complete" && "Complete Event"}
              {actionType === "delete" && "Delete Event"}
            </DialogTitle>
            <DialogDescription>
              {actionType === "delete"
                ? "This action cannot be undone. The event will be permanently deleted."
                : `This will ${actionType} the event: ${selectedEvent?.title}`}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">Reason</label>
              <Textarea
                placeholder="Enter reason for this action..."
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
                required
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowActionDialog(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={executeEventAction}
              variant={
                actionType === "delete" || actionType === "cancel"
                  ? "destructive"
                  : "default"
              }
              disabled={!actionReason.trim()}
            >
              {actionType === "publish" && "Publish Event"}
              {actionType === "cancel" && "Cancel Event"}
              {actionType === "complete" && "Complete Event"}
              {actionType === "delete" && "Delete Event"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Bulk Action Dialog */}
      <Dialog open={showBulkDialog} onOpenChange={setShowBulkDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Bulk Action Confirmation</DialogTitle>
            <DialogDescription>
              You are about to {bulkAction} {selectedIds.length} event(s).
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">Reason</label>
              <Textarea
                placeholder="Enter reason for this bulk action..."
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
                required
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBulkDialog(false)}>
              Cancel
            </Button>
            <Button
              onClick={executeBulkAction}
              variant={
                bulkAction === "delete" || bulkAction === "cancel"
                  ? "destructive"
                  : "default"
              }
              disabled={!actionReason.trim()}
            >
              Execute {bulkAction}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SidebarProvider>
  );
}

export default withAuth(EventsPage);
