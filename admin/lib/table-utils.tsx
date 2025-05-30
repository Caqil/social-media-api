// lib/table-utils.tsx - Fixed version with proper filter values
import React from "react";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  IconUserCheck,
  IconUserX,
  IconBan,
  IconCheck,
  IconShield,
  IconUsers,
  IconMail,
  IconCalendar,
  IconEye,
  IconEyeOff,
  IconHeart,
  IconMessage,
  IconShare,
  IconFlag,
  IconAlertTriangle,
  IconClock,
  IconMapPin,
  IconLink,
  IconPhoto,
  IconVideo,
  IconFile,
  IconMusic,
} from "@tabler/icons-react";

// Common cell formatters
export const TableFormatters = {
  // User avatar and name formatter
  userProfile: (user: any, showEmail = false) => (
    <div className="flex items-center gap-3">
      <Avatar className="h-8 w-8">
        <AvatarImage src={user.profile_picture} alt={user.username} />
        <AvatarFallback>
          {(user.first_name?.[0] || user.username[0] || "").toUpperCase()}
        </AvatarFallback>
      </Avatar>
      <div className="min-w-0">
        <div className="font-medium flex items-center gap-2">
          {user.first_name
            ? `${user.first_name} ${user.last_name || ""}`.trim()
            : user.username}
          {user.is_verified && (
            <IconUserCheck className="h-4 w-4 text-blue-600" />
          )}
        </div>
        <div className="text-sm text-muted-foreground">
          @{user.username}
          {showEmail && user.email && (
            <span className="ml-2">• {user.email}</span>
          )}
        </div>
      </div>
    </div>
  ),

  // Email formatter with icon
  email: (email: string) => (
    <div className="flex items-center gap-2">
      <IconMail className="h-4 w-4 text-muted-foreground" />
      <span className="text-sm">{email}</span>
    </div>
  ),

  // Role formatter with colors and icons
  role: (role: string) => {
    const roleConfig = {
      user: {
        color: "bg-gray-100 text-gray-800 border-gray-200",
        icon: IconUsers,
      },
      moderator: {
        color: "bg-blue-100 text-blue-800 border-blue-200",
        icon: IconShield,
      },
      admin: {
        color: "bg-purple-100 text-purple-800 border-purple-200",
        icon: IconShield,
      },
      super_admin: {
        color: "bg-red-100 text-red-800 border-red-200",
        icon: IconShield,
      },
    };

    const config =
      roleConfig[role as keyof typeof roleConfig] || roleConfig.user;
    const Icon = config.icon;

    return (
      <Badge variant="outline" className={config.color}>
        <Icon className="h-3 w-3 mr-1" />
        {role.replace("_", " ").toUpperCase()}
      </Badge>
    );
  },

  // Verification status formatter
  verification: (isVerified: boolean) => (
    <Badge variant={isVerified ? "default" : "secondary"}>
      {isVerified ? (
        <>
          <IconUserCheck className="h-3 w-3 mr-1" />
          Verified
        </>
      ) : (
        <>
          <IconUserX className="h-3 w-3 mr-1" />
          Unverified
        </>
      )}
    </Badge>
  ),

  // User status (active/suspended/inactive)
  userStatus: (user: any) => {
    if (user.is_suspended) {
      return (
        <Badge variant="destructive">
          <IconBan className="h-3 w-3 mr-1" />
          Suspended
        </Badge>
      );
    }
    return (
      <Badge variant={user.is_active ? "default" : "secondary"}>
        {user.is_active ? (
          <>
            <IconCheck className="h-3 w-3 mr-1" />
            Active
          </>
        ) : (
          <>
            <IconBan className="h-3 w-3 mr-1" />
            Inactive
          </>
        )}
      </Badge>
    );
  },

  // Post visibility formatter
  visibility: (visibility: string) => {
    const visibilityConfig = {
      public: {
        color: "bg-green-100 text-green-800 border-green-200",
        icon: IconEye,
      },
      friends: {
        color: "bg-blue-100 text-blue-800 border-blue-200",
        icon: IconUsers,
      },
      private: {
        color: "bg-gray-100 text-gray-800 border-gray-200",
        icon: IconEyeOff,
      },
    };

    const config =
      visibilityConfig[visibility as keyof typeof visibilityConfig] ||
      visibilityConfig.public;
    const Icon = config.icon;

    return (
      <Badge variant="outline" className={config.color}>
        <Icon className="h-3 w-3 mr-1" />
        {visibility.charAt(0).toUpperCase() + visibility.slice(1)}
      </Badge>
    );
  },

  // Generic status formatter
  status: (status: string) => {
    const statusConfig = {
      active: {
        color: "bg-green-100 text-green-800 border-green-200",
        icon: IconCheck,
      },
      inactive: {
        color: "bg-gray-100 text-gray-800 border-gray-200",
        icon: IconBan,
      },
      pending: {
        color: "bg-yellow-100 text-yellow-800 border-yellow-200",
        icon: IconClock,
      },
      suspended: {
        color: "bg-red-100 text-red-800 border-red-200",
        icon: IconBan,
      },
      resolved: {
        color: "bg-blue-100 text-blue-800 border-blue-200",
        icon: IconCheck,
      },
      reviewing: {
        color: "bg-purple-100 text-purple-800 border-purple-200",
        icon: IconEye,
      },
      rejected: {
        color: "bg-orange-100 text-orange-800 border-orange-200",
        icon: IconAlertTriangle,
      },
      published: {
        color: "bg-green-100 text-green-800 border-green-200",
        icon: IconCheck,
      },
      draft: {
        color: "bg-gray-100 text-gray-800 border-gray-200",
        icon: IconEyeOff,
      },
      cancelled: {
        color: "bg-red-100 text-red-800 border-red-200",
        icon: IconBan,
      },
      completed: {
        color: "bg-blue-100 text-blue-800 border-blue-200",
        icon: IconCheck,
      },
    };

    const config =
      statusConfig[status.toLowerCase() as keyof typeof statusConfig];
    if (!config) {
      return <Badge variant="outline">{status}</Badge>;
    }

    const Icon = config.icon;

    return (
      <Badge variant="outline" className={config.color}>
        <Icon className="h-3 w-3 mr-1" />
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </Badge>
    );
  },

  // Date formatter with relative time
  date: (dateString: string, showRelative = true) => {
    if (!dateString) return <span className="text-muted-foreground">-</span>;

    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    let relativeText = "";
    if (showRelative) {
      if (diffDays === 0) relativeText = "Today";
      else if (diffDays === 1) relativeText = "Yesterday";
      else if (diffDays < 7) relativeText = `${diffDays} days ago`;
      else if (diffDays < 30)
        relativeText = `${Math.floor(diffDays / 7)} weeks ago`;
      else if (diffDays < 365)
        relativeText = `${Math.floor(diffDays / 30)} months ago`;
      else relativeText = `${Math.floor(diffDays / 365)} years ago`;
    }

    return (
      <div className="flex items-center gap-2 text-sm">
        <IconCalendar className="h-4 w-4 text-muted-foreground" />
        <div>
          <div>{date.toLocaleDateString()}</div>
          {showRelative && relativeText && (
            <div className="text-xs text-muted-foreground">{relativeText}</div>
          )}
        </div>
      </div>
    );
  },

  // Engagement stats formatter
  engagement: (stats: {
    likes?: number;
    comments?: number;
    shares?: number;
    views?: number;
  }) => (
    <div className="text-sm space-y-1">
      <div className="flex items-center gap-4">
        {stats.likes !== undefined && (
          <span className="flex items-center gap-1">
            <IconHeart className="h-3 w-3" />
            {stats.likes.toLocaleString()}
          </span>
        )}
        {stats.comments !== undefined && (
          <span className="flex items-center gap-1">
            <IconMessage className="h-3 w-3" />
            {stats.comments.toLocaleString()}
          </span>
        )}
        {stats.shares !== undefined && (
          <span className="flex items-center gap-1">
            <IconShare className="h-3 w-3" />
            {stats.shares.toLocaleString()}
          </span>
        )}
      </div>
      {stats.views !== undefined && (
        <div className="text-muted-foreground flex items-center gap-1">
          <IconEye className="h-3 w-3" />
          {stats.views.toLocaleString()} views
        </div>
      )}
    </div>
  ),

  // User stats formatter
  userStats: (user: any) => (
    <div className="text-sm space-y-1">
      <div className="flex items-center gap-4">
        <span>{user.posts_count?.toLocaleString() || 0} posts</span>
        <span>{user.followers_count?.toLocaleString() || 0} followers</span>
      </div>
      <div className="text-muted-foreground">
        {user.following_count?.toLocaleString() || 0} following
      </div>
    </div>
  ),

  // Media type formatter
  mediaType: (type: string, url?: string) => {
    const typeConfig = {
      image: {
        color: "bg-green-100 text-green-800 border-green-200",
        icon: IconPhoto,
      },
      video: {
        color: "bg-blue-100 text-blue-800 border-blue-200",
        icon: IconVideo,
      },
      audio: {
        color: "bg-purple-100 text-purple-800 border-purple-200",
        icon: IconMusic,
      },
      document: {
        color: "bg-orange-100 text-orange-800 border-orange-200",
        icon: IconFile,
      },
    };

    const config =
      typeConfig[type as keyof typeof typeConfig] || typeConfig.document;
    const Icon = config.icon;

    return (
      <div className="flex items-center gap-2">
        <Badge variant="outline" className={config.color}>
          <Icon className="h-3 w-3 mr-1" />
          {type.charAt(0).toUpperCase() + type.slice(1)}
        </Badge>
        {url && (
          <Button variant="ghost" size="sm" asChild>
            <a href={url} target="_blank" rel="noopener noreferrer">
              <IconLink className="h-3 w-3" />
            </a>
          </Button>
        )}
      </div>
    );
  },

  // Boolean formatter with custom labels
  boolean: (value: boolean, trueLabel = "Yes", falseLabel = "No") => (
    <Badge variant={value ? "default" : "secondary"}>
      {value ? trueLabel : falseLabel}
    </Badge>
  ),

  // Number formatter with locale
  number: (value: number, options?: Intl.NumberFormatOptions) => (
    <span className="tabular-nums">
      {value.toLocaleString("en-US", options)}
    </span>
  ),

  // Currency formatter
  currency: (value: number, currency = "USD") => (
    <span className="tabular-nums">
      {new Intl.NumberFormat("en-US", {
        style: "currency",
        currency,
      }).format(value)}
    </span>
  ),

  // Location formatter
  location: (location: string) => (
    <div className="flex items-center gap-2 text-sm">
      <IconMapPin className="h-4 w-4 text-muted-foreground" />
      <span>{location}</span>
    </div>
  ),

  // Priority formatter
  priority: (priority: string) => {
    const priorityConfig = {
      low: { color: "bg-green-100 text-green-800 border-green-200" },
      medium: { color: "bg-yellow-100 text-yellow-800 border-yellow-200" },
      high: { color: "bg-orange-100 text-orange-800 border-orange-200" },
      critical: { color: "bg-red-100 text-red-800 border-red-200" },
    };

    const config =
      priorityConfig[priority.toLowerCase() as keyof typeof priorityConfig];
    if (!config) {
      return <Badge variant="outline">{priority}</Badge>;
    }

    return (
      <Badge variant="outline" className={config.color}>
        {priority.charAt(0).toUpperCase() + priority.slice(1)}
      </Badge>
    );
  },

  // Report reason formatter
  reportReason: (reason: string) => (
    <div className="flex items-center gap-2">
      <IconFlag className="h-4 w-4 text-red-500" />
      <span className="capitalize">{reason.replace(/_/g, " ")}</span>
    </div>
  ),

  // Truncated text with tooltip
  truncatedText: (text: string, maxLength = 50) => {
    if (!text) return <span className="text-muted-foreground">-</span>;

    if (text.length <= maxLength) {
      return <span>{text}</span>;
    }

    return <span title={text}>{text.substring(0, maxLength)}...</span>;
  },

  // Content preview formatter
  contentPreview: (content: string, maxLength = 100) => {
    if (!content)
      return <span className="text-muted-foreground">No content</span>;

    const truncated =
      content.length > maxLength
        ? content.substring(0, maxLength) + "..."
        : content;

    return (
      <div className="max-w-xs">
        <span className="text-sm" title={content}>
          {truncated}
        </span>
      </div>
    );
  },
};

// Common bulk actions by entity type
export const BulkActions = {
  user: [
    { label: "Verify Selected", action: "verify", icon: IconUserCheck },
    { label: "Unverify Selected", action: "unverify", icon: IconUserX },
    { label: "Activate Selected", action: "activate", icon: IconCheck },
    { label: "Deactivate Selected", action: "deactivate", icon: IconBan },
    {
      label: "Suspend Selected",
      action: "suspend",
      icon: IconBan,
      variant: "destructive" as const,
    },
    {
      label: "Delete Selected",
      action: "delete",
      icon: IconBan,
      variant: "destructive" as const,
    },
  ],

  post: [
    { label: "Hide Selected", action: "hide", icon: IconEyeOff },
    { label: "Show Selected", action: "show", icon: IconEye },
    {
      label: "Delete Selected",
      action: "delete",
      icon: IconBan,
      variant: "destructive" as const,
    },
  ],

  comment: [
    { label: "Hide Selected", action: "hide", icon: IconEyeOff },
    { label: "Show Selected", action: "show", icon: IconEye },
    {
      label: "Delete Selected",
      action: "delete",
      icon: IconBan,
      variant: "destructive" as const,
    },
  ],

  report: [
    { label: "Resolve Selected", action: "resolve", icon: IconCheck },
    { label: "Reject Selected", action: "reject", icon: IconAlertTriangle },
    { label: "Mark as Reviewing", action: "reviewing", icon: IconEye },
  ],

  group: [
    { label: "Activate Selected", action: "activate", icon: IconCheck },
    { label: "Deactivate Selected", action: "deactivate", icon: IconBan },
    {
      label: "Delete Selected",
      action: "delete",
      icon: IconBan,
      variant: "destructive" as const,
    },
  ],

  event: [
    { label: "Publish Selected", action: "publish", icon: IconCheck },
    {
      label: "Cancel Selected",
      action: "cancel",
      icon: IconBan,
      variant: "destructive" as const,
    },
    {
      label: "Delete Selected",
      action: "delete",
      icon: IconBan,
      variant: "destructive" as const,
    },
  ],
};

// ✅ FIXED: Common filter options by entity type - changed empty strings to "all"
export const FilterOptions = {
  user: {
    is_verified: [
      { label: "All", value: "all", icon: IconEye }, // ✅ Changed from "" to "all"
      { label: "Verified", value: "true", icon: IconUserCheck },
      { label: "Unverified", value: "false", icon: IconUserX },
    ],
    is_active: [
      { label: "All", value: "all", icon: IconEye }, // ✅ Changed from "" to "all"
      { label: "Active", value: "true", icon: IconCheck },
      { label: "Inactive", value: "false", icon: IconBan },
    ],
    is_suspended: [
      { label: "All", value: "all", icon: IconEye }, // ✅ Changed from "" to "all"
      { label: "Not Suspended", value: "false", icon: IconCheck },
      { label: "Suspended", value: "true", icon: IconAlertTriangle },
    ],
    role: [
      { label: "All Roles", value: "all", icon: IconEye }, // ✅ Changed from "" to "all"
      { label: "User", value: "user", icon: IconUsers },
      { label: "Moderator", value: "moderator", icon: IconShield },
      { label: "Admin", value: "admin", icon: IconShield },
      { label: "Super Admin", value: "super_admin", icon: IconShield },
    ],
  },

  post: {
    is_reported: [
      { label: "All", value: "all", icon: IconEye }, // ✅ Changed from "" to "all"
      { label: "Not Reported", value: "false", icon: IconCheck },
      { label: "Reported", value: "true", icon: IconFlag },
    ],
    is_hidden: [
      { label: "All", value: "all", icon: IconEye }, // ✅ Changed from "" to "all"
      { label: "Visible", value: "false", icon: IconEye },
      { label: "Hidden", value: "true", icon: IconEyeOff },
    ],
    visibility: [
      { label: "All", value: "all", icon: IconEye }, // ✅ Changed from "" to "all"
      { label: "Public", value: "public", icon: IconEye },
      { label: "Friends", value: "friends", icon: IconUsers },
      { label: "Private", value: "private", icon: IconEyeOff },
    ],
  },

  report: {
    status: [
      { label: "All", value: "all", icon: IconEye }, // ✅ Changed from "" to "all"
      { label: "Pending", value: "pending", icon: IconClock },
      { label: "Reviewing", value: "reviewing", icon: IconEye },
      { label: "Resolved", value: "resolved", icon: IconCheck },
      { label: "Rejected", value: "rejected", icon: IconAlertTriangle },
    ],
    priority: [
      { label: "All", value: "all", icon: IconEye }, // ✅ Changed from "" to "all"
      { label: "Low", value: "low", icon: IconCheck },
      { label: "Medium", value: "medium", icon: IconClock },
      { label: "High", value: "high", icon: IconAlertTriangle },
      { label: "Critical", value: "critical", icon: IconFlag },
    ],
  },
};

// Utility function to get default columns for entity types
export const getDefaultColumns = (entityType: string): any[] => {
  // This would return default column configurations for different entity types
  // Implementation would depend on your specific needs
  return [];
};

export default TableFormatters;
