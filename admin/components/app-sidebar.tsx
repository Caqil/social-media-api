// components/app-sidebar.tsx - Complete Admin Navigation
"use client";

import * as React from "react";
import {
  IconCamera,
  IconChartBar,
  IconDashboard,
  IconDatabase,
  IconFileAi,
  IconFileDescription,
  IconFileWord,
  IconFolder,
  IconHelp,
  IconInnerShadowTop,
  IconListDetails,
  IconReport,
  IconSearch,
  IconSettings,
  IconUsers,
  IconMessages,
  IconUsersGroup,
  IconCalendarEvent,
  IconPhoto,
  IconMail,
  IconHeart,
  IconHash,
  IconAt,
  IconCloudUpload,
  IconBell,
  IconFlag,
  IconShield,
  IconTools,
  IconChevronDown,
  IconChevronRight,
} from "@tabler/icons-react";

import { NavDocuments } from "@/components/nav-documents";
import { NavMain } from "@/components/nav-main";
import { NavSecondary } from "@/components/nav-secondary";
import { NavUser } from "@/components/nav-user";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from "@/components/ui/sidebar";
import { useAuth } from "@/contexts/auth-context";
import { MonitorCheck } from "lucide-react";

const data = {
  user: {
    name: "Admin User",
    email: "admin@example.com",
    avatar: "/avatars/admin.jpg",
  },

  // Main Navigation
  navMain: [
    {
      title: "Dashboard",
      url: "/admin/dashboard",
      icon: IconDashboard,
    },
    {
      title: "Analytics",
      url: "/admin/analytics",
      icon: IconChartBar,
    },
  ],

  // Content Management
  contentManagement: [
    {
      title: "Users",
      icon: IconUsers,
      url: "/admin/users",
    },
    {
      title: "Posts",
      icon: IconMessages,
      url: "/admin/posts",
      items: [
        {
          title: "All Posts",
          url: "/admin/posts",
        },
        {
          title: "Reported Posts",
          url: "/admin/posts?is_reported=true",
        },
        {
          title: "Hidden Posts",
          url: "/admin/posts?is_hidden=true",
        },
      ],
    },
    {
      title: "Comments",
      icon: IconFileDescription,
      url: "/admin/comments",
      items: [
        {
          title: "All Comments",
          url: "/admin/comments",
        },
        {
          title: "Reported Comments",
          url: "/admin/comments?is_reported=true",
        },
        {
          title: "Hidden Comments",
          url: "/admin/comments?is_hidden=true",
        },
      ],
    },
    {
      title: "Groups",
      icon: IconUsersGroup,
      url: "/admin/groups",
      items: [
        {
          title: "All Groups",
          url: "/admin/groups",
        },
        {
          title: "Public Groups",
          url: "/admin/groups?type=public",
        },
        {
          title: "Private Groups",
          url: "/admin/groups?type=private",
        },
      ],
    },
    {
      title: "Events",
      icon: IconCalendarEvent,
      url: "/admin/events",
      items: [
        {
          title: "All Events",
          url: "/admin/events",
        },
        {
          title: "Upcoming Events",
          url: "/admin/events?status=published",
        },
        {
          title: "Draft Events",
          url: "/admin/events?status=draft",
        },
      ],
    },
    {
      title: "Stories",
      icon: IconPhoto,
      url: "/admin/stories",
      items: [
        {
          title: "All Stories",
          url: "/admin/stories",
        },
        {
          title: "Hidden Stories",
          url: "/admin/stories?is_hidden=true",
        },
      ],
    },
    {
      title: "Messages",
      icon: IconMail,
      url: "/admin/messages",
      items: [
        {
          title: "All Messages",
          url: "/admin/messages",
        },
        {
          title: "Conversations",
          url: "/admin/messages/conversations",
        },
      ],
    },
  ],

  // Engagement & Social
  engagement: [
    {
      title: "Reports",
      icon: IconReport,
      url: "/admin/reports",
      items: [
        {
          title: "All Reports",
          url: "/admin/reports",
        },
        {
          title: "Pending Reports",
          url: "/admin/reports?status=pending",
        },
        {
          title: "Resolved Reports",
          url: "/admin/reports?status=resolved",
        },
      ],
    },
    {
      title: "Follows",
      icon: IconHeart,
      url: "/admin/follows",
      items: [
        {
          title: "All Follows",
          url: "/admin/follows",
        },
        {
          title: "Relationships",
          url: "/admin/follows/relationships",
        },
      ],
    },
    {
      title: "Likes",
      icon: IconHeart,
      url: "/admin/likes",
      items: [
        {
          title: "All Likes",
          url: "/admin/likes",
        },
        {
          title: "Like Statistics",
          url: "/admin/likes/stats",
        },
      ],
    },
    {
      title: "Hashtags",
      icon: IconHash,
      url: "/admin/hashtags",
      items: [
        {
          title: "All Hashtags",
          url: "/admin/hashtags",
        },
        {
          title: "Trending Hashtags",
          url: "/admin/hashtags/trending",
        },
        {
          title: "Blocked Hashtags",
          url: "/admin/hashtags?is_blocked=true",
        },
      ],
    },
    {
      title: "Mentions",
      icon: IconAt,
      url: "/admin/mentions",
    },
  ],

  // Media & Content
  media: [
    {
      title: "Media",
      icon: IconCloudUpload,
      url: "/admin/media",
      items: [
        {
          title: "All Media",
          url: "/admin/media",
        },
        {
          title: "Images",
          url: "/admin/media?type=image",
        },
        {
          title: "Videos",
          url: "/admin/media?type=video",
        },
        {
          title: "Storage Stats",
          url: "/admin/media/storage",
        },
      ],
    },
    {
      title: "Notifications",
      icon: IconBell,
      url: "/admin/notifications",
      items: [
        {
          title: "All Notifications",
          url: "/admin/notifications",
        },
        {
          title: "Send Notification",
          url: "/admin/notifications/send",
        },
        {
          title: "Broadcast",
          url: "/admin/notifications/broadcast",
        },
      ],
    },
  ],

  // System & Configuration
  system: [
    {
      title: "System",
      icon: MonitorCheck,
      url: "/admin/system",
      items: [
        {
          title: "Health Status",
          url: "/admin/system/health",
        },
        {
          title: "Performance",
          url: "/admin/system/performance",
        },
        {
          title: "Logs",
          url: "/admin/system/logs",
        },
        {
          title: "Database",
          url: "/admin/system/database",
        },
      ],
    },
    {
      title: "Configuration",
      icon: IconSettings,
      url: "/admin/config",
      items: [
        {
          title: "General Settings",
          url: "/admin/config",
        },
        {
          title: "Feature Flags",
          url: "/admin/config/features",
        },
        {
          title: "Rate Limits",
          url: "/admin/config/rate-limits",
        },
      ],
    },
  ],

  // Secondary Navigation
  navSecondary: [
    {
      title: "Help & Support",
      url: "/admin/help",
      icon: IconHelp,
    },
    {
      title: "Search",
      url: "/admin/search",
      icon: IconSearch,
    },
  ],
};

// Navigation Group Component
function NavGroup({
  title,
  items,
  defaultOpen = false,
}: {
  title: string;
  items: any[];
  defaultOpen?: boolean;
}) {
  const [isOpen, setIsOpen] = React.useState(defaultOpen);

  return (
    <Collapsible
      open={isOpen}
      onOpenChange={setIsOpen}
      className="group/collapsible"
    >
      <CollapsibleTrigger asChild>
        <SidebarMenuButton className="w-full justify-between">
          <span className="font-medium text-sidebar-foreground/70">
            {title}
          </span>
          {isOpen ? (
            <IconChevronDown className="ml-auto transition-transform" />
          ) : (
            <IconChevronRight className="ml-auto transition-transform" />
          )}
        </SidebarMenuButton>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <SidebarMenu>
          {items.map((item) => (
            <SidebarMenuItem key={item.title}>
              {item.items ? (
                <Collapsible className="group/nested">
                  <CollapsibleTrigger asChild>
                    <SidebarMenuButton className="w-full">
                      {item.icon && <item.icon className="mr-2 h-4 w-4" />}
                      <span>{item.title}</span>
                      <IconChevronRight className="ml-auto transition-transform group-data-[state=open]/nested:rotate-90" />
                    </SidebarMenuButton>
                  </CollapsibleTrigger>
                  <CollapsibleContent>
                    <SidebarMenuSub>
                      {item.items.map((subItem: any) => (
                        <SidebarMenuSubItem key={subItem.title}>
                          <SidebarMenuSubButton asChild>
                            <a href={subItem.url}>
                              <span>{subItem.title}</span>
                            </a>
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                      ))}
                    </SidebarMenuSub>
                  </CollapsibleContent>
                </Collapsible>
              ) : (
                <SidebarMenuButton asChild>
                  <a href={item.url}>
                    {item.icon && <item.icon className="mr-2 h-4 w-4" />}
                    <span>{item.title}</span>
                  </a>
                </SidebarMenuButton>
              )}
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </CollapsibleContent>
    </Collapsible>
  );
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { user } = useAuth();

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:!p-1.5"
            >
              <a href="/admin/dashboard">
                <IconInnerShadowTop className="!size-5" />
                <span className="text-base font-semibold">Admin Panel</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent className="gap-0">
        {/* Main Navigation */}
        <div className="p-2">
          <NavMain items={data.navMain} />
        </div>

        {/* Content Management */}
        <div className="border-t border-sidebar-border p-2">
          <NavGroup
            title="Content Management"
            items={data.contentManagement}
            defaultOpen={true}
          />
        </div>

        {/* Engagement & Social */}
        <div className="border-t border-sidebar-border p-2">
          <NavGroup title="Engagement & Social" items={data.engagement} />
        </div>

        {/* Media & Content */}
        <div className="border-t border-sidebar-border p-2">
          <NavGroup title="Media & Content" items={data.media} />
        </div>

        {/* System & Configuration */}
        <div className="border-t border-sidebar-border p-2">
          <NavGroup title="System & Configuration" items={data.system} />
        </div>

        {/* Secondary Navigation */}
        <div className="mt-auto border-t border-sidebar-border p-2">
          <NavSecondary items={data.navSecondary} />
        </div>
      </SidebarContent>

      <SidebarFooter>
        <NavUser
          user={{
            name: user?.username || "Admin User",
            email: user?.email || "admin@example.com",
            avatar: "/avatars/admin.jpg",
          }}
        />
      </SidebarFooter>
    </Sidebar>
  );
}

// Export navigation data for use in other components
export { data as navigationData };
